package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"personal-disk/config"
	"personal-disk/model"
	"strings"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

const uploadRequestOverhead int64 = 10 << 20

// 上传文件
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// 获取配置
	cfg := config.MustGetConfig()

	// 1. 解析请求
	// 限制请求方法
	if r.Method != http.MethodPost {
		http.Error(w, "提交方式错误", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, cfg.Upload.MaxSize+uploadRequestOverhead)

	// 设置缓冲区大小（使用配置中的最大上传大小）
	err := r.ParseMultipartForm(cfg.Upload.MaxSize)
	if err != nil {
		http.Error(w, "解析表单失败: "+err.Error(), http.StatusBadRequest)
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "没有上传文件", http.StatusBadRequest)
		return
	}

	// 2. 向数据库添加文件信息
	var response Response
	statusCode := http.StatusOK
	for _, fileHeader := range files {
		filename, err := safeFileName(fileHeader.Filename)
		if err != nil {
			response.Error = fmt.Sprintf("文件名无效: %s", fileHeader.Filename)
			statusCode = http.StatusBadRequest
			break
		}

		// 检查文件类型是否允许
		ext := strings.ToLower(filepath.Ext(filename))
		if ext != "" {
			ext = ext[1:] // 移除点号
		}
		if !cfg.IsAllowedFileType(ext) {
			response.Error = fmt.Sprintf("不允许上传 %s 类型的文件", ext)
			statusCode = http.StatusBadRequest
			break
		}

		// 检查文件大小
		if fileHeader.Size > cfg.Upload.MaxSize {
			response.Error = fmt.Sprintf("文件 %s 大小超过限制（%d MB）",
				filename, cfg.Upload.MaxSize/(1024*1024))
			statusCode = http.StatusRequestEntityTooLarge
			break
		}

		exists, err := model.JudgeFileExists(filename)
		if err != nil {
			response.Error = fmt.Sprintf("检查文件是否存在失败: %v", err)
			statusCode = http.StatusInternalServerError
			break
		}
		if exists {
			response.Error = fmt.Sprintf("文件 %s 已存在", filename)
			statusCode = http.StatusConflict
			break
		}

		// 打开上传的文件
		srcFile, err := fileHeader.Open() // 使用里面的临时指针
		if err != nil {
			response.Error = fmt.Sprintf("打开文件 %s 失败: %v", filename, err)
			statusCode = http.StatusInternalServerError
			break
		}

		size := fileHeader.Size
		destPath, err := uploadFilePath(cfg.Upload.Directory, filename)
		if err != nil {
			srcFile.Close()
			response.Error = err.Error()
			statusCode = http.StatusBadRequest
			break
		}

		// 确保上传目录存在
		if err := os.MkdirAll(cfg.Upload.Directory, 0755); err != nil {
			srcFile.Close()
			response.Error = fmt.Sprintf("创建上传目录失败: %v", err)
			statusCode = http.StatusInternalServerError
			break
		}

		dstFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			srcFile.Close()
			if os.IsExist(err) {
				response.Error = fmt.Sprintf("文件 %s 已存在", filename)
				statusCode = http.StatusConflict
				break
			}
			response.Error = fmt.Sprintf("创建文件 %s 失败: %v", destPath, err)
			statusCode = http.StatusInternalServerError
			break
		}

		_, err = io.Copy(dstFile, srcFile)
		closeErr := dstFile.Close()
		srcFile.Close()
		if err != nil {
			response.Error = fmt.Sprintf("保存文件 %s 失败: %v", filename, err)
			// 删除可能已部分写入的文件
			os.Remove(destPath)
			statusCode = http.StatusInternalServerError
			break
		}
		if closeErr != nil {
			response.Error = fmt.Sprintf("保存文件 %s 失败: %v", filename, closeErr)
			os.Remove(destPath)
			statusCode = http.StatusInternalServerError
			break
		}

		// 将文件信息添加到数据库
		err = model.AddFileInfo(filename, destPath, size)
		if err != nil {
			response.Error = fmt.Sprintf("添加文件信息失败: %v", err)
			os.Remove(destPath) // 删除已保存的文件
			statusCode = http.StatusInternalServerError
			break
		}
	}
	// 3. 根据结果返回 JSON
	w.Header().Set("Content-Type", "application/json")

	if response.Error != "" {
		response.Success = false
		w.WriteHeader(statusCode)
	} else {
		response.Success = true
		response.Message = "文件上传成功"
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(response)
}

// 获取文件列表
func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "提交方式错误", http.StatusMethodNotAllowed)
		return
	}

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
	if r.Method != http.MethodGet {
		http.Error(w, "提交方式错误", http.StatusMethodNotAllowed)
		return
	}

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
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(fileInfo.FileName)))
	w.Header().Set("Content-Type", "application/octet-stream")

	file, err := os.Open(fileInfo.FilePath)
	if err != nil {
		http.Error(w, "打开文件失败", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		http.Error(w, "获取文件信息失败", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, fileInfo.FileName, fi.ModTime(), file)
}

// 删除文件
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "提交方式错误", http.StatusMethodNotAllowed)
		return
	}

	// 1. 获取文件相关数据
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 || parts[1] != "delete" {
		http.Error(w, "无效的请求路径", http.StatusBadRequest)
		return
	}
	id := parts[2]

	fileInfo, err := model.GetFileInfoByID(id)
	if err != nil {
		http.Error(w, "查询文件信息失败", http.StatusInternalServerError)
		return
	}

	// 2. 删除实际文件
	err = os.Remove(fileInfo.FilePath)
	if err != nil && !os.IsNotExist(err) {
		http.Error(w, "删除文件失败", http.StatusInternalServerError)
		return
	}

	// 3. 删除数据库中的文件信息
	err = model.DeleteFileInfoByID(id)
	if err != nil {
		http.Error(w, "删除文件信息失败", http.StatusInternalServerError)
		return
	}

	response := Response{
		Success: true,
		Message: "文件删除成功",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 重命名文件
func RenameHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "提交方式错误", http.StatusMethodNotAllowed)
		return
	}

	// 获取配置
	cfg := config.MustGetConfig()

	// 1. 获取文件相关数据
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 || parts[1] != "rename" {
		http.Error(w, "无效的请求路径", http.StatusBadRequest)
		return
	}
	id := parts[2]

	oldFile, err := model.GetFileInfoByID(id)
	if err != nil {
		http.Error(w, "获取文件信息失败", http.StatusInternalServerError)
		return
	}

	var data struct {
		NewName string `json:"file_name"`
	}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		// 处理错误（例如返回 400）
		http.Error(w, "无效的 JSON", http.StatusBadRequest)
		return
	}
	newName, err := safeFileName(data.NewName)
	if err != nil {
		http.Error(w, "文件名无效", http.StatusBadRequest)
		return
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(newName)), ".")
	if !cfg.IsAllowedFileType(ext) {
		http.Error(w, "不允许使用该文件类型", http.StatusBadRequest)
		return
	}

	if newName == oldFile.FileName {
		response := Response{
			Success: true,
			Message: "文件重命名成功",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	exists, err := model.JudgeFileExists(newName)
	if err != nil {
		http.Error(w, "检查文件是否存在失败", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "文件已存在", http.StatusConflict)
		return
	}

	newPath, err := uploadFilePath(cfg.Upload.Directory, newName)
	if err != nil {
		http.Error(w, "文件名无效", http.StatusBadRequest)
		return
	}

	err = os.Rename(oldFile.FilePath, newPath)
	if err != nil {
		http.Error(w, "重命名文件失败", http.StatusInternalServerError)
		return
	}

	err = model.RenameFileInfoByID(id, newName, newPath)
	if err != nil {
		_ = os.Rename(newPath, oldFile.FilePath)
		http.Error(w, "重命名文件信息失败", http.StatusInternalServerError)
		return
	}

	response := Response{
		Success: true,
		Message: "文件重命名成功",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 返回 admin 页面
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	// 获取配置
	cfg := config.MustGetConfig()

	// 检查是否已登录
	if !IsLoggedIn(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	templatePath := filepath.Join(cfg.Templates.Directory, "admin.html")
	http.ServeFile(w, r, templatePath)
}

// 返回 index 页面
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// 获取配置
	cfg := config.MustGetConfig()

	templatePath := filepath.Join(cfg.Templates.Directory, "index.html")
	http.ServeFile(w, r, templatePath)
}

func safeFileName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		return "", fmt.Errorf("文件名不能为空")
	}
	if strings.ContainsRune(name, 0) || strings.ContainsAny(name, `/\`) || filepath.Base(name) != name {
		return "", fmt.Errorf("文件名不能包含路径分隔符")
	}
	return name, nil
}

func uploadFilePath(uploadDir, filename string) (string, error) {
	cleanDir := filepath.Clean(uploadDir)
	destPath := filepath.Join(cleanDir, filename)
	rel, err := filepath.Rel(cleanDir, destPath)
	if err != nil {
		return "", fmt.Errorf("生成文件路径失败: %v", err)
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("文件路径超出上传目录")
	}
	return destPath, nil
}
