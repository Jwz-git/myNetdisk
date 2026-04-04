package main

import (
	"fmt"
	"log"
	"net/http"
	"personal-disk/config"
	"personal-disk/controller"
	"personal-disk/model"
)

func main() {
	// 初始化配置
	cfg := config.InitConfig()

	// 初始化数据库
	if err := initDatabase(cfg); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer closeDB()

	// 启动HTTP服务器
	startServer(cfg)
}

// initDatabase 初始化数据库连接
func initDatabase(cfg *config.Config) error {
	dsn := cfg.GetDSN()

	err := model.InitDB(dsn)
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	fmt.Printf("数据库连接成功: %s@%s:%s/%s\n",
		cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)

	return nil
}

// closeDB 关闭数据库连接
func closeDB() {
	if model.DB != nil {
		model.DB.Close()
		log.Println("数据库连接已关闭")
	}
}

// startServer 启动HTTP服务器
func startServer(cfg *config.Config) {
	// 设置路由
	setupRoutes(cfg)

	// 服务器地址
	addr := cfg.GetServerAddr()

	// 输出启动信息
	fmt.Printf("=== %s ===\n", cfg.App.Name)
	fmt.Printf("版本: %s\n", cfg.App.Version)
	fmt.Printf("环境: %s\n", config.GetEnv("APP_ENV", "development"))
	fmt.Printf("服务器启动: http://%s\n", addr)
	fmt.Printf("管理页面: http://%s/admin\n", addr)
	fmt.Printf("当前时间: %s\n", model.GetCurrentTime())

	// 启动服务器
	log.Fatal(http.ListenAndServe(addr, nil))
}

// setupRoutes 设置路由
func setupRoutes(cfg *config.Config) {
	// 静态文件服务
	http.Handle(cfg.Static.URLPrefix,
		http.StripPrefix(cfg.Static.URLPrefix,
			http.FileServer(http.Dir(cfg.Static.Directory))))

	// API 路由
	http.HandleFunc("/api/upload", controller.UploadHandler)
	http.HandleFunc("/api/files", controller.ListFilesHandler)
	http.HandleFunc("/api/download/", controller.DownloadHandler)
	http.HandleFunc("/api/delete/", controller.DeleteHandler)
	http.HandleFunc("/api/rename/", controller.RenameHandler)
	http.HandleFunc("/api/login", controller.LoginHandler)

	// 页面路由
	http.HandleFunc("/admin", controller.AdminHandler)
	http.HandleFunc("/login", controller.LoginPageHandler)
	http.HandleFunc("/", controller.IndexHandler)
}