package controller

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"personal-disk/internal/config"
	"personal-disk/internal/model"
	"strings"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

const uploadRequestOverhead int64 = 10 << 20

type uploadEntry struct {
	Name  string
	Path  string
	Size  int64
	IsDir bool
}

type uploadFileItem struct {
	Header       *multipart.FileHeader
	RelativePath string
	DestPath     string
}

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

	// 2. 校验上传内容并整理为顶层文件/文件夹条目
	var response Response
	statusCode := http.StatusOK
	fail := func(status int, format string, args ...any) {
		response.Error = fmt.Sprintf(format, args...)
		statusCode = status
	}

	formPaths := r.MultipartForm.Value["paths"]
	entriesByName := make(map[string]*uploadEntry)
	entryOrder := make([]*uploadEntry, 0)
	uploadItems := make([]uploadFileItem, 0, len(files))
	seenDestPaths := make(map[string]struct{})

	for i, fileHeader := range files {
		rawPath := fileHeader.Filename
		if i < len(formPaths) && strings.TrimSpace(formPaths[i]) != "" {
			rawPath = formPaths[i]
		}

		relativePath, err := safeUploadRelativePath(rawPath)
		if err != nil {
			fail(http.StatusBadRequest, "文件路径无效: %s", rawPath)
			break
		}

		// 检查文件类型是否允许
		ext := strings.ToLower(path.Ext(relativePath))
		if ext != "" {
			ext = ext[1:] // 移除点号
		}
		if !cfg.IsAllowedFileType(ext) {
			fail(http.StatusBadRequest, "不允许上传 %s 类型的文件", ext)
			break
		}

		// 检查文件大小
		if fileHeader.Size > cfg.Upload.MaxSize {
			fail(http.StatusRequestEntityTooLarge, "文件 %s 大小超过限制（%d MB）",
				relativePath, cfg.Upload.MaxSize/(1024*1024))
			break
		}

		destPath, err := uploadRelativePath(cfg.Upload.Directory, relativePath)
		if err != nil {
			fail(http.StatusBadRequest, err.Error())
			break
		}
		if _, exists := seenDestPaths[destPath]; exists {
			fail(http.StatusBadRequest, "上传内容包含重复路径: %s", relativePath)
			break
		}
		seenDestPaths[destPath] = struct{}{}

		parts := strings.Split(relativePath, "/")
		entryName := parts[0]
		isDir := len(parts) > 1
		entryPath, err := uploadRelativePath(cfg.Upload.Directory, entryName)
		if err != nil {
			fail(http.StatusBadRequest, err.Error())
			break
		}

		entry, exists := entriesByName[entryName]
		if !exists {
			entry = &uploadEntry{
				Name:  entryName,
				Path:  entryPath,
				IsDir: isDir,
			}
			entriesByName[entryName] = entry
			entryOrder = append(entryOrder, entry)
		} else if entry.IsDir != isDir {
			fail(http.StatusBadRequest, "上传内容中同时包含同名文件和文件夹: %s", entryName)
			break
		} else if !isDir {
			fail(http.StatusBadRequest, "上传内容包含重复文件: %s", entryName)
			break
		}

		if isDir {
			entry.Size += fileHeader.Size
		} else {
			entry.Size = fileHeader.Size
		}
		uploadItems = append(uploadItems, uploadFileItem{
			Header:       fileHeader,
			RelativePath: relativePath,
			DestPath:     destPath,
		})
	}

	if response.Error == "" {
		for _, entry := range entryOrder {
			exists, err := model.JudgeFileExists(entry.Name)
			if err != nil {
				fail(http.StatusInternalServerError, "检查文件是否存在失败: %v", err)
				break
			}
			if exists {
				fail(http.StatusConflict, "%s %s 已存在", entryKindName(entry.IsDir), entry.Name)
				break
			}
			if _, err := os.Stat(entry.Path); err == nil {
				fail(http.StatusConflict, "%s %s 已存在", entryKindName(entry.IsDir), entry.Name)
				break
			} else if !os.IsNotExist(err) {
				fail(http.StatusInternalServerError, "检查上传路径失败: %v", err)
				break
			}
		}
	}

	if response.Error == "" {
		// 确保上传目录存在
		if err := os.MkdirAll(cfg.Upload.Directory, 0755); err != nil {
			fail(http.StatusInternalServerError, "创建上传目录失败: %v", err)
		}
	}

	if response.Error == "" {
		for _, item := range uploadItems {
			// 打开上传的文件
			srcFile, err := item.Header.Open() // 使用里面的临时指针
			if err != nil {
				cleanupUploadedEntries(entryOrder)
				fail(http.StatusInternalServerError, "打开文件 %s 失败: %v", item.RelativePath, err)
				break
			}

			if err := os.MkdirAll(filepath.Dir(item.DestPath), 0755); err != nil {
				srcFile.Close()
				cleanupUploadedEntries(entryOrder)
				fail(http.StatusInternalServerError, "创建上传目录失败: %v", err)
				break
			}

			dstFile, err := os.OpenFile(item.DestPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
			if err != nil {
				srcFile.Close()
				cleanupUploadedEntries(entryOrder)
				if os.IsExist(err) {
					fail(http.StatusConflict, "文件 %s 已存在", item.RelativePath)
					break
				}
				fail(http.StatusInternalServerError, "创建文件 %s 失败: %v", item.DestPath, err)
				break
			}

			_, err = io.Copy(dstFile, srcFile)
			closeErr := dstFile.Close()
			srcFile.Close()
			if err != nil {
				cleanupUploadedEntries(entryOrder)
				fail(http.StatusInternalServerError, "保存文件 %s 失败: %v", item.RelativePath, err)
				break
			}
			if closeErr != nil {
				cleanupUploadedEntries(entryOrder)
				fail(http.StatusInternalServerError, "保存文件 %s 失败: %v", item.RelativePath, closeErr)
				break
			}
		}
	}

	if response.Error == "" {
		addedNames := make([]string, 0, len(entryOrder))
		for _, entry := range entryOrder {
			// 将文件/文件夹信息添加到数据库
			err := model.AddFileEntryInfo(entry.Name, entry.Path, entry.Size, entry.IsDir)
			if err != nil {
				for _, name := range addedNames {
					_ = model.DeleteFileInfoByName(name)
				}
				cleanupUploadedEntries(entryOrder)
				fail(http.StatusInternalServerError, "添加文件信息失败: %v", err)
				break
			}
			addedNames = append(addedNames, entry.Name)
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

func cleanupUploadedEntries(entries []*uploadEntry) {
	for _, entry := range entries {
		if entry.Path != "" {
			_ = os.RemoveAll(entry.Path)
		}
	}
}

func entryKindName(isDir bool) string {
	if isDir {
		return "文件夹"
	}
	return "文件"
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
	cfg := config.MustGetConfig()

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
	if !pathWithinUploadDir(cfg.Upload.Directory, fileInfo.FilePath) {
		http.Error(w, "文件路径无效", http.StatusBadRequest)
		return
	}

	info, err := os.Stat(fileInfo.FilePath)
	if err != nil {
		http.Error(w, "获取文件信息失败", http.StatusInternalServerError)
		return
	}

	if fileInfo.IsDir {
		if !info.IsDir() {
			http.Error(w, "文件夹不存在", http.StatusInternalServerError)
			return
		}
		zipName := fileInfo.FileName + ".zip"
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(zipName)))
		w.Header().Set("Content-Type", "application/zip")
		if err := streamDirectoryAsZip(w, fileInfo.FilePath, fileInfo.FileName); err != nil {
			fmt.Printf("压缩文件夹失败: %v\n", err)
		}
		return
	}
	if info.IsDir() {
		http.Error(w, "文件类型无效", http.StatusInternalServerError)
		return
	}

	file, err := os.Open(fileInfo.FilePath)
	if err != nil {
		http.Error(w, "打开文件失败", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 2. 下载文件
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", url.PathEscape(fileInfo.FileName)))
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeContent(w, r, fileInfo.FileName, info.ModTime(), file)
}

// 删除文件
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "提交方式错误", http.StatusMethodNotAllowed)
		return
	}
	cfg := config.MustGetConfig()

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
	if !pathWithinUploadDir(cfg.Upload.Directory, fileInfo.FilePath) {
		http.Error(w, "文件路径无效", http.StatusBadRequest)
		return
	}

	// 2. 删除实际文件
	if fileInfo.IsDir {
		err = os.RemoveAll(fileInfo.FilePath)
	} else {
		err = os.Remove(fileInfo.FilePath)
	}
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
	if !pathWithinUploadDir(cfg.Upload.Directory, oldFile.FilePath) {
		http.Error(w, "文件路径无效", http.StatusBadRequest)
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

	if !oldFile.IsDir {
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(newName)), ".")
		if !cfg.IsAllowedFileType(ext) {
			http.Error(w, "不允许使用该文件类型", http.StatusBadRequest)
			return
		}
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
	if _, err := os.Stat(newPath); err == nil {
		http.Error(w, "文件已存在", http.StatusConflict)
		return
	} else if !os.IsNotExist(err) {
		http.Error(w, "检查文件路径失败", http.StatusInternalServerError)
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

func safeUploadRelativePath(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || strings.ContainsRune(name, 0) || strings.Contains(name, "\\") {
		return "", fmt.Errorf("文件路径无效")
	}

	segments := strings.Split(name, "/")
	cleanSegments := make([]string, 0, len(segments))
	for _, segment := range segments {
		cleanSegment, err := safeFileName(segment)
		if err != nil {
			return "", err
		}
		cleanSegments = append(cleanSegments, cleanSegment)
	}

	relativePath := strings.Join(cleanSegments, "/")
	if relativePath == "." || relativePath == ".." || strings.HasPrefix(relativePath, "../") || path.IsAbs(relativePath) {
		return "", fmt.Errorf("文件路径超出上传目录")
	}
	if path.Clean(relativePath) != relativePath {
		return "", fmt.Errorf("文件路径无效")
	}
	return relativePath, nil
}

func uploadFilePath(uploadDir, filename string) (string, error) {
	filename, err := safeFileName(filename)
	if err != nil {
		return "", err
	}
	return uploadRelativePath(uploadDir, filename)
}

func uploadRelativePath(uploadDir, relativePath string) (string, error) {
	cleanDir := filepath.Clean(uploadDir)
	destPath := filepath.Join(cleanDir, filepath.FromSlash(relativePath))
	rel, err := filepath.Rel(cleanDir, destPath)
	if err != nil {
		return "", fmt.Errorf("生成文件路径失败: %v", err)
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("文件路径超出上传目录")
	}
	return destPath, nil
}

func pathWithinUploadDir(uploadDir, targetPath string) bool {
	baseAbs, err := filepath.Abs(filepath.Clean(uploadDir))
	if err != nil {
		return false
	}
	targetAbs, err := filepath.Abs(filepath.Clean(targetPath))
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(baseAbs, targetAbs)
	if err != nil {
		return false
	}
	return rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

func streamDirectoryAsZip(w io.Writer, dirPath, rootName string) error {
	rootName, err := safeFileName(rootName)
	if err != nil {
		return err
	}

	zipWriter := zip.NewWriter(w)
	rootInfo, err := os.Stat(dirPath)
	if err != nil {
		_ = zipWriter.Close()
		return err
	}

	rootHeader, err := zip.FileInfoHeader(rootInfo)
	if err != nil {
		_ = zipWriter.Close()
		return err
	}
	rootHeader.Name = rootName + "/"
	if _, err := zipWriter.CreateHeader(rootHeader); err != nil {
		_ = zipWriter.Close()
		return err
	}

	walkErr := filepath.WalkDir(dirPath, func(currentPath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if currentPath == dirPath {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(dirPath, currentPath)
		if err != nil {
			return err
		}
		zipPath := path.Join(rootName, filepath.ToSlash(relPath))
		if entry.IsDir() {
			zipPath += "/"
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = zipPath
		if !entry.IsDir() {
			header.Method = zip.Deflate
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		file, err := os.Open(currentPath)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(writer, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})

	closeErr := zipWriter.Close()
	if walkErr != nil {
		return walkErr
	}
	return closeErr
}
