package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/yuma25/uni-steps/domain"
	"github.com/yuma25/uni-steps/port"
	"golang.org/x/oauth2"
)

// AuthHandler は Google OAuth 2.0 による認証を管理するハンドラーである．
type AuthHandler struct {
	userRepo port.UserRepository // ユーザー情報を保存するためのリポジトリである．
	oauthCfg *oauth2.Config      // Google OAuth 2.0 の設定情報である．
}

// NewAuthHandler は AuthHandler を初期化し，ルーティングを登録する．
func NewAuthHandler(e *echo.Echo, ur port.UserRepository, cfg *oauth2.Config) {
	h := &AuthHandler{
		userRepo: ur,
		oauthCfg: cfg,
	}
	e.GET("/api/auth/google/login", h.GoogleLogin)
	e.GET("/api/auth/google/callback", h.GoogleCallback)
}

// getOAuthConfig はホスト名に応じて適切な RedirectURL を設定した OAuth 設定を返す．
// 開発環境（localhost）からのリクエスト時は自動でローカル用 URL を割り当てる．
func (h *AuthHandler) getOAuthConfig(c echo.Context) oauth2.Config {
	cfg := *h.oauthCfg
	host := c.Request().Host
	if strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") {
		cfg.RedirectURL = fmt.Sprintf("http://%s/api/auth/google/callback", host)
	}
	return cfg
}

// GoogleLogin は Google の認可画面へリダイレクトする．
func (h *AuthHandler) GoogleLogin(c echo.Context) error {
	// CSRF 対策としてランダムな文字列（state）を生成する．
	state := uuid.New().String()

	isSecure := os.Getenv("GO_ENV") == "production" || c.Scheme() == "https" || c.Request().Header.Get("X-Forwarded-Proto") == "https"

	// state を Cookie に保存する（ブラウザ側で保持させ，Callback 時に検証する）．
	cookie := &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(15 * time.Minute), // 15 分間有効
		HttpOnly: true,                             // JavaScript からアクセス不可
		Secure:   isSecure,                         // HTTPS または本番環境で true
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	cfg := h.getOAuthConfig(c)
	url := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback は Google からの認可コードを受け取り，ユーザー情報の取得・保存・ログインを行う．
func (h *AuthHandler) GoogleCallback(c echo.Context) error {
	// state の検証を行う（CSRF 対策）．
	state := c.QueryParam("state")
	cookie, err := c.Cookie("oauth_state")
	if err != nil {
		log.Printf("[AuthHandler] Cookie取得エラー (oauth_state が存在しない/ブロックされた): %v", err)
		return c.JSON(http.StatusForbidden, map[string]string{"error": "不正なリクエスト（state 不一致: Cookieが送信されませんでした）が検出された"})
	}
	if state == "" || state != cookie.Value {
		log.Printf("[AuthHandler] state 不一致: queryState=%s, cookieState=%s", state, cookie.Value)
		return c.JSON(http.StatusForbidden, map[string]string{"error": "不正なリクエスト（state 不一致）が検出された"})
	}

	// 使用済みの Cookie を削除する．
	c.SetCookie(&http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   c.Scheme() == "https" || c.Request().Header.Get("X-Forwarded-Proto") == "https" || os.Getenv("GO_ENV") == "production",
		SameSite: http.SameSiteLaxMode,
	})

	code := c.QueryParam("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "認可コードが提供されていない"})
	}

	cfg := h.getOAuthConfig(c)

	// 認可コードをアクセストークンと交換する．
	token, err := cfg.Exchange(c.Request().Context(), code)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "トークンの交換に失敗した"})
	}

	// Google の UserInfo API からユーザーのプロフィールを取得する．
	client := cfg.Client(c.Request().Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ユーザー情報の取得に失敗した"})
	}
	defer resp.Body.Close()

	var googleUser struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ユーザー情報の解析に失敗した"})
	}

	// 3．メールアドレスを元に，既存ユーザーか新規ユーザーかを判定する．
	user, err := h.userRepo.FindByEmail(c.Request().Context(), googleUser.Email)
	if err != nil {
		log.Printf("[Auth] FindByEmail エラー (email: %s): %v\n", googleUser.Email, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "データベース検索に失敗した"})
	}

	if user == nil {
		// 新規ユーザー登録（サインアップ）
		user = &domain.User{
			ID:    uuid.New().String(),
			Name:  googleUser.Name,
			Email: googleUser.Email,
		}
	}

	// トークン情報を更新して保存する．
	user.GoogleAccessToken = token.AccessToken
	if token.RefreshToken != "" {
		user.GoogleRefreshToken = token.RefreshToken
	}
	user.GoogleTokenExpiry = token.Expiry

	if err := h.userRepo.Save(c.Request().Context(), user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "ユーザー情報の保存に失敗した"})
	}

	// ダッシュボードへリダイレクトする．
	host := c.Request().Host
	frontendUrl := os.Getenv("FRONTEND_URL")
	if strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") || frontendUrl == "" {
		frontendUrl = "http://localhost:5173"
	}
	redirectUrl := frontendUrl + "/select-group?user_id=" + user.ID
	return c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}
