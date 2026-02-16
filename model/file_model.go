package model

import (
    "time"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // 导入 MySQL 驱动，下划线表示只执行 init 函数
	"log"
)

type FileInfo struct {
	ID         int          `json:"id"`
	FileName   string       `json:"file_name"`
	FilePath   string       `json:"file_path"`
	FileSize   int64        `json:"file_size"`
	UploadTime time.Time    `json:"upload_time"`
}

// 全局数据库连接对象，方便其他函数使用
var DB *sql.DB

// 初始化数据库连接
func InitDB(dataSourceName string) error {

	var err error

	// 打开数据库连接（只是初始化连接池）
	DB, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %v", err)
	}

	// 设置连接池参数
	DB.SetMaxOpenConns(10) // 最大打开连接数
	DB.SetMaxIdleConns(5)  // 最大空闲连接数

	// 尝试与数据库建立连接，验证 DSN 是否正确
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("连接数据库失败: %v", err)
	}

	log.Println("数据库连接成功")
	return nil
}

func GetAllFileInfos() ([]FileInfo, error) {
    cursor, err := DB.Query("SELECT id, file_name, file_path, file_size, upload_time FROM file_info")
    if err != nil{
        fmt.Printf("查询文件信息失败: %v\n", err)
        return nil, err
    }
    defer cursor.Close()

    fileinfos := make([]FileInfo,0)

    for cursor.Next(){
        var f FileInfo
        err := cursor.Scan(&f.ID, &f.FileName, &f.FilePath, &f.FileSize, &f.UploadTime)
        if err != nil{
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

func GetFileInfoByID(id string) (FileInfo,error){
	var f FileInfo
	cursor,err := DB.Query("SELECT id, file_name, file_path, file_size, upload_time FROM file_info WHERE id = ?", id)
	if err != nil{
		fmt.Printf("查询文件信息失败: %v\n", err)
		return FileInfo{}, err
	}
	defer cursor.Close()
	if cursor.Next() {
		err := cursor.Scan(&f.ID, &f.FileName, &f.FilePath, &f.FileSize, &f.UploadTime)
		if err != nil {
			fmt.Printf("扫描文件信息失败: %v\n", err)
			return FileInfo{}, err
		}
		return f, nil
	}
	return FileInfo{}, fmt.Errorf("未找到ID为 %s 的文件信息", id)
}

func AddFileInfo(fileName, filePath string, fileSize int64) error {
    _,err := DB.Exec("INSERT INTO file_info (file_name, file_path, file_size, upload_time) VALUES (?, ?, ?, ?)", fileName, filePath, fileSize, time.Now())
    return err
}
