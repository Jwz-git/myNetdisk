package model

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

type FileInfo struct {
	ID         int       `json:"id"`
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"file_path"`
	FileSize   int64     `json:"file_size"`
	IsDir      bool      `json:"is_dir"`
	UpdateTime time.Time `json:"update_time"`
}

// 全局数据库连接对象，方便其他函数使用
var DB *sql.DB
var databaseDriver string

// 初始化数据库连接
func InitDB(driver, dataSourceName string) error {
	if driver != "mysql" && driver != "sqlite" {
		return fmt.Errorf("不支持的数据库驱动 %q", driver)
	}
	if driver == "sqlite" {
		if err := ensureSQLiteDirectory(dataSourceName); err != nil {
			return err
		}
	}

	var err error

	// 打开数据库连接（只是初始化连接池）
	DB, err = sql.Open(driver, dataSourceName)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}

	// 设置连接池参数
	if driver == "sqlite" {
		DB.SetMaxOpenConns(1)
		DB.SetMaxIdleConns(1)
	} else {
		DB.SetMaxOpenConns(10)
		DB.SetMaxIdleConns(5)
	}

	// 尝试与数据库建立连接，验证 DSN 是否正确
	if err = DB.Ping(); err != nil {
		_ = DB.Close()
		DB = nil
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	// 初始化数据库表
	databaseDriver = driver
	if err = initTables(); err != nil {
		_ = DB.Close()
		DB = nil
		return fmt.Errorf("初始化数据库表失败: %v", err)
	}

	log.Println("数据库连接成功")
	return nil
}

func ensureSQLiteDirectory(dsn string) error {
	if strings.Contains(dsn, ":memory:") {
		return nil
	}
	path := strings.TrimPrefix(strings.SplitN(dsn, "?", 2)[0], "file:")
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建 SQLite 数据目录失败: %v", err)
	}
	return nil
}

// 初始化数据库表
func initTables() error {
	if databaseDriver == "sqlite" {
		return initSQLiteTables()
	}
	return initMySQLTables()
}

func initMySQLTables() error {
	// 创建 file_info 表
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS file_info (
		id INT AUTO_INCREMENT PRIMARY KEY,
		file_name VARCHAR(255) NOT NULL,
		file_path VARCHAR(500) NOT NULL,
		file_size BIGINT,
		is_dir BOOLEAN NOT NULL DEFAULT FALSE,
		update_time DATETIME DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`

	_, err := DB.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("创建 file_info 表失败: %v", err)
	}
	if err := ensureFileInfoColumns(); err != nil {
		return err
	}

	log.Println("数据库表初始化成功")
	return nil
}

func initSQLiteTables() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS file_info (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_name TEXT NOT NULL,
		file_path TEXT NOT NULL,
		file_size INTEGER,
		is_dir BOOLEAN NOT NULL DEFAULT FALSE,
		update_time DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := DB.Exec(createTableSQL); err != nil {
		return fmt.Errorf("创建 file_info 表失败: %v", err)
	}
	if err := ensureSQLiteFileInfoColumns(); err != nil {
		return err
	}
	log.Println("数据库表初始化成功")
	return nil
}

func ensureSQLiteFileInfoColumns() error {
	rows, err := DB.Query("PRAGMA table_info(file_info)")
	if err != nil {
		return fmt.Errorf("检查 file_info 表结构失败: %v", err)
	}
	defer rows.Close()

	hasIsDir := false
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return fmt.Errorf("读取 file_info 表结构失败: %v", err)
		}
		if name == "is_dir" {
			hasIsDir = true
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("遍历 file_info 表结构失败: %v", err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("关闭 file_info 表结构结果集失败: %v", err)
	}
	if hasIsDir {
		return nil
	}
	if _, err := DB.Exec("ALTER TABLE file_info ADD COLUMN is_dir BOOLEAN NOT NULL DEFAULT FALSE"); err != nil {
		return fmt.Errorf("添加 file_info.is_dir 字段失败: %v", err)
	}
	return nil
}

func ensureFileInfoColumns() error {
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
			AND TABLE_NAME = 'file_info'
			AND COLUMN_NAME = 'is_dir'
	`).Scan(&count)
	if err != nil {
		return fmt.Errorf("检查 file_info.is_dir 字段失败: %v", err)
	}
	if count > 0 {
		return nil
	}

	_, err = DB.Exec("ALTER TABLE file_info ADD COLUMN is_dir BOOLEAN NOT NULL DEFAULT FALSE AFTER file_size")
	if err != nil {
		return fmt.Errorf("添加 file_info.is_dir 字段失败: %v", err)
	}
	return nil
}

func GetAllFileInfos() ([]FileInfo, error) {
	cursor, err := DB.Query("SELECT id, file_name, file_path, file_size, is_dir, update_time FROM file_info ORDER BY update_time DESC")
	if err != nil {
		fmt.Printf("查询文件信息失败: %v\n", err)
		return nil, err
	}
	defer cursor.Close()

	fileinfos := make([]FileInfo, 0)

	for cursor.Next() {
		var f FileInfo
		err := cursor.Scan(&f.ID, &f.FileName, &f.FilePath, &f.FileSize, &f.IsDir, &f.UpdateTime)
		if err != nil {
			fmt.Printf("扫描文件信息失败: %v\n", err)
			return nil, err
		}
		fileinfos = append(fileinfos, f)
	}

	if err = cursor.Err(); err != nil {
		fmt.Printf("遍历结果集失败: %v\n", err)
		return nil, fmt.Errorf("遍历结果集错误: %v", err)
	}

	return fileinfos, nil
}

func GetFileInfoByID(id string) (FileInfo, error) {
	var f FileInfo
	cursor, err := DB.Query("SELECT id, file_name, file_path, file_size, is_dir, update_time FROM file_info WHERE id = ?", id)
	if err != nil {
		fmt.Printf("查询文件信息失败: %v\n", err)
		return FileInfo{}, err
	}
	defer cursor.Close()
	if cursor.Next() {
		err := cursor.Scan(&f.ID, &f.FileName, &f.FilePath, &f.FileSize, &f.IsDir, &f.UpdateTime)
		if err != nil {
			fmt.Printf("扫描文件信息失败: %v\n", err)
			return FileInfo{}, err
		}
		return f, nil
	}
	return FileInfo{}, fmt.Errorf("未找到ID为 %s 的文件信息", id)
}

func JudgeFileExists(fileName string) (bool, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM file_info WHERE file_name = ?", fileName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("查询文件信息失败: %v", err)
	}
	return count > 0, nil
}

func AddFileInfo(fileName, filePath string, fileSize int64) error {
	return AddFileEntryInfo(fileName, filePath, fileSize, false)
}

func AddFileEntryInfo(fileName, filePath string, fileSize int64, isDir bool) error {
	_, err := DB.Exec("INSERT INTO file_info (file_name, file_path, file_size, is_dir, update_time) VALUES (?, ?, ?, ?, ?)", fileName, filePath, fileSize, isDir, time.Now())
	return err
}

func DeleteFileInfoByID(id string) error {
	_, err := DB.Exec("DELETE FROM file_info WHERE id = ?", id)
	return err
}

func DeleteFileInfoByName(fileName string) error {
	_, err := DB.Exec("DELETE FROM file_info WHERE file_name = ?", fileName)
	return err
}

func RenameFileInfoByID(id, newName, newPath string) error {
	_, err := DB.Exec("UPDATE file_info SET file_name = ?, file_path = ?, update_time = ? WHERE id = ?", newName, newPath, time.Now(), id)
	return err
}

func GetCurrentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
