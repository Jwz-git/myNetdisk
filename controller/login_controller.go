package controller

import (
	"encoding/json"
	"net/http"
	"personal-disk/config"
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

	// 获取配置
	cfg := config.MustGetConfig()

	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	// 验证用户名和密码（使用配置中的凭据）
	if req.Username != cfg.Admin.Username || req.Password != cfg.Admin.Password {
		response := Response{
			Success: false,
			Error:   "用户名或密码错误",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 创建会话（使用配置中的会话设置）
	session, err := r.Cookie(cfg.Session.CookieName)
	if err != nil {
		session = &http.Cookie{
			Name:   cfg.Session.CookieName,
			Value:  cfg.Session.CookieValue,
			Path:   cfg.Session.CookiePath,
			MaxAge: cfg.Session.MaxAge,
		}
	}
	session.Value = cfg.Session.CookieValue
	session.Path = cfg.Session.CookiePath
	session.MaxAge = cfg.Session.MaxAge
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
	// 获取配置
	cfg := config.MustGetConfig()

	// 删除会话
	session := &http.Cookie{
		Name:   cfg.Session.CookieName,
		Value:  "",
		Path:   cfg.Session.CookiePath,
		MaxAge: -1, // 立即删除
	}
	http.SetCookie(w, session)

	// 重定向到登录页面
	http.Redirect(w, r, "/login", http.StatusFound)
}

// 检查是否已登录
func IsLoggedIn(r *http.Request) bool {
	cfg := config.MustGetConfig()

	session, err := r.Cookie(cfg.Session.CookieName)
	if err != nil {
		return false
	}
	return session.Value == cfg.Session.CookieValue
}

// 登录页面处理函数
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	// 获取配置
	cfg := config.MustGetConfig()

	// 如果已经登录，直接重定向到 admin 页面
	if IsLoggedIn(r) {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	templatePath := cfg.Templates.Directory + "/login.html"
	http.ServeFile(w, r, templatePath)
}