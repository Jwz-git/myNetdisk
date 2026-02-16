package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"personal-disk/model"
	"strings"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// 上传文件
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析请求
	// 限制请求方法
	if r.Method != http.MethodPost {
		http.Error(w, "提交方式错误", http.StatusMethodNotAllowed)
		return
	}
	// 设置缓冲区大小
	err := r.ParseMultipartForm(512 << 20) // 512 MB
	if err != nil {
		http.Error(w, "解析表单失败: "+err.Error(), http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "没有上传文件", http.StatusBadRequest)
		return
	}

	// 2. 向数据库添加文件信息
	var response Response
	for _, fileHeader := range files {
		// 打开上传的文件
		srcFile, err := fileHeader.Open() // 使用里面的临时指针
		if err != nil {
			response.Error = fmt.Sprintf("打开文件 %s 失败: %v", fileHeader.Filename, err)
			break
		}
		defer srcFile.Close()

		filename := fileHeader.Filename
		size := fileHeader.Size
		destPath := filepath.Join("uploads", filename)

		dstFile, err := os.Create(destPath)
		if err != nil {
			response.Error = fmt.Sprintf("创建文件 %s 失败: %v", destPath, err)
			break
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			response.Error = fmt.Sprintf("保存文件 %s 失败: %v", fileHeader.Filename, err)
			// 删除可能已部分写入的文件
			os.Remove(destPath)
			break
		}

		// 将文件信息添加到数据库
		err = model.AddFileInfo(filename, destPath, size)
		if err != nil {
			response.Error = fmt.Sprintf("添加文件信息失败: %v", err)
			os.Remove(destPath) // 删除已保存的文件
			break
		}
	}
	// 3. 根据结果返回 JSON
	w.Header().Set("Content-Type", "application/json")

	if response.Error != "" {
		response.Success = false
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Success = true
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(response)
}

// 获取文件列表
func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 获取数据库中的文件信息列表
	files, err := model.GetAllFileInfos()
	if err != nil {
		http.Error(w, "查询文件信息失败", http.StatusInternalServerError)
		return
	}
	jsonData, err := json.Marshal(files)
	if err != nil {
		http.Error(w, "编码文件信息失败", http.StatusInternalServerError)
		return
	}
	// 2. 返回 JSON 数据
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// 下载文件
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 获取文件相关数据
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 || parts[1] != "download" {
		http.Error(w, "无效的请求路径", http.StatusBadRequest)
		return
	}
	id := parts[2]

	fileInfo, err := model.GetFileInfoByID(id)
	if err != nil {
		http.Error(w, "查询文件信息失败", http.StatusInternalServerError)
		return
	}
	// fmt.Printf("文件名：%s, 路径：%s\n", fileInfo.FileName, fileInfo.FilePath)

	// 2. 下载文件
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=utf-8''%s", fileInfo.FileName))
    w.Header().Set("Content-Type", "application/octet-stream")
    http.ServeFile(w, r, fileInfo.FilePath)
}

// 返回 admin 页面
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/admin.html")
}

// 返回 index 页面
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("访问了 index 页面")
	http.ServeFile(w, r, "templates/index.html")
}
