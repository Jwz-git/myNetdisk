package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"personal-disk/controller"
	"personal-disk/model"
)

func connectDB() error {
	// 从环境变量读取数据库配置
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "personal_disk")

	// 构建 MySQL 数据源名称
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// 初始化数据库连接
	err := model.InitDB(dsn)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err) // 如果失败，直接终止程序
	}

	// 连接成功
	fmt.Println("数据库连接成功，准备启动 HTTP 服务...")

	return nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func closeDB() {
	if model.DB != nil {
		model.DB.Close()
		log.Println("数据库连接已关闭")
	}
}

func main() {
	connectDB()

	fmt.Println("time:", model.GetCurrentTime())
	fmt.Println("HTTP 服务正在运行，访问 http://localhost:8080 来使用个人网盘")

	// 静态文件
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// API
	http.HandleFunc("/api/upload", controller.UploadHandler)
	http.HandleFunc("/api/files", controller.ListFilesHandler)
	http.HandleFunc("/api/download/", controller.DownloadHandler)
	http.HandleFunc("/api/delete/", controller.DeleteHandler)
	http.HandleFunc("/api/rename/", controller.RenameHandler)
	http.HandleFunc("/api/login", controller.LoginHandler)

	// 页面
	http.HandleFunc("/admin", controller.AdminHandler)
	http.HandleFunc("/login", controller.LoginPageHandler)
	http.HandleFunc("/", controller.IndexHandler)

	http.ListenAndServe(":8080", nil)

	// 程序退出时关闭数据库连接
	defer closeDB()
}
