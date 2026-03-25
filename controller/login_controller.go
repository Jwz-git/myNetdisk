package controller

import (
	"encoding/json"
	"net/http"
)

// 硬编码的管理员凭据（实际项目中应该从数据库或配置文件中读取）
const (
	AdminUsername = "admin"
	AdminPassword = "123456"
)

// 登录请求结构
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// 登录处理函数
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	// 验证用户名和密码
	if req.Username != AdminUsername || req.Password != AdminPassword {
		response := Response{
			Success: false,
			Error:   "用户名或密码错误",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 创建会话
	session, err := r.Cookie("admin_session")
	if err != nil {
		session = &http.Cookie{
			Name:   "admin_session",
			Value:  "admin_logged_in",
			Path:   "/",
			MaxAge: 86400, // 24小时
		}
	}
	session.Value = "admin_logged_in"
	session.Path = "/"
	session.MaxAge = 86400
	http.SetCookie(w, session)

	response := Response{
		Success: true,
		Message: "登录成功",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// 退出登录处理函数
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// 删除会话
	session := &http.Cookie{
		Name:   "admin_session",
		Value:  "",
		Path:   "/",
		MaxAge: -1, // 立即删除
	}
	http.SetCookie(w, session)

	// 重定向到登录页面
	http.Redirect(w, r, "/login", http.StatusFound)
}

// 检查是否已登录
func IsLoggedIn(r *http.Request) bool {
	session, err := r.Cookie("admin_session")
	if err != nil {
		return false
	}
	return session.Value == "admin_logged_in"
}

// 登录页面处理函数
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	// 如果已经登录，直接重定向到 admin 页面
	if IsLoggedIn(r) {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}
	http.ServeFile(w, r, "templates/login.html")
}