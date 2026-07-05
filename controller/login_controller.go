package controller

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"personal-disk/config"
	"strconv"
	"strings"
	"time"
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

	token, err := newSessionToken(cfg)
	if err != nil {
		http.Error(w, "创建会话失败", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, sessionCookie(cfg, token, r))

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
		Name:     cfg.Session.CookieName,
		Value:    "",
		Path:     cfg.Session.CookiePath,
		MaxAge:   -1, // 立即删除
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isHTTPS(r),
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
	return verifySessionToken(cfg, session.Value)
}

// RequireLogin 为需要管理员权限的 API 增加登录校验。
func RequireLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsLoggedIn(r) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Response{
				Success: false,
				Error:   "请先登录",
			})
			return
		}
		next(w, r)
	}
}

func sessionCookie(cfg *config.Config, value string, r *http.Request) *http.Cookie {
	return &http.Cookie{
		Name:     cfg.Session.CookieName,
		Value:    value,
		Path:     cfg.Session.CookiePath,
		MaxAge:   cfg.Session.MaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isHTTPS(r),
	}
}

func newSessionToken(cfg *config.Config) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(time.Duration(cfg.Session.MaxAge) * time.Second).Unix()
	payload := fmt.Sprintf("%s|%d|%s",
		cfg.Admin.Username,
		expiresAt,
		base64.RawURLEncoding.EncodeToString(nonce),
	)
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))
	return encodedPayload + "." + signSessionPayload(cfg, encodedPayload), nil
}

func verifySessionToken(cfg *config.Config, token string) bool {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}

	expectedSignature := signSessionPayload(cfg, parts[0])
	if !hmac.Equal([]byte(parts[1]), []byte(expectedSignature)) {
		return false
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false
	}

	payloadParts := strings.Split(string(payloadBytes), "|")
	if len(payloadParts) != 3 || payloadParts[0] != cfg.Admin.Username {
		return false
	}

	expiresAt, err := strconv.ParseInt(payloadParts[1], 10, 64)
	if err != nil {
		return false
	}

	return time.Now().Unix() <= expiresAt
}

func signSessionPayload(cfg *config.Config, encodedPayload string) string {
	key := []byte(cfg.Admin.Password + "\x00" + cfg.Session.CookieValue)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(encodedPayload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func isHTTPS(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
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
