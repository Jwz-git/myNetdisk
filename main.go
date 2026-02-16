package main

import (
	"fmt"
	"log"
    "net/http"
    "personal-disk/controller"
	"personal-disk/model"
)

func connectDB() error {
	 // 定义 MySQL 数据源名称
    dsn := "root:@tcp(127.0.0.1:3306)/personal_disk?charset=utf8mb4&parseTime=True&loc=Local"

    // 初始化数据库连接
    err := model.InitDB(dsn)
    if err != nil {
        log.Fatalf("数据库初始化失败: %v", err) // 如果失败，直接终止程序
    }

    // 连接成功
    fmt.Println("数据库连接成功，准备启动 HTTP 服务...")

    return nil
}

func closeDB() {
    if model.DB != nil {
        model.DB.Close()
        log.Println("数据库连接已关闭")
    }
}

func main(){
    connectDB()

    fmt.Println("开始")

    // 静态文件
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    // API
    http.HandleFunc("/api/upload", controller.UploadHandler)
    http.HandleFunc("/api/files", controller.ListFilesHandler)
    http.HandleFunc("/api/download/", controller.DownloadHandler)

    // 页面
    http.HandleFunc("/admin", controller.AdminHandler)
    http.HandleFunc("/", controller.IndexHandler)

    http.ListenAndServe(":8080", nil)

    // 程序退出时关闭数据库连接
    defer closeDB()
}