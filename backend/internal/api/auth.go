package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/remotescope/backend/internal/auth"
	"github.com/remotescope/backend/internal/models"
	"github.com/remotescope/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo   *repository.UserRepository
	jwtSvc     *auth.JWTService
	oauthSvc   *auth.OAuthService
	bcryptCost int
}

func NewAuthHandler(
	userRepo *repository.UserRepository,
	jwtSvc *auth.JWTService,
	oauthSvc *auth.OAuthService,
	bcryptCost int,
) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtSvc:     jwtSvc,
		oauthSvc:   oauthSvc,
		bcryptCost: bcryptCost,
	}
}

// ──────────────────────────────────────────
// POST /api/v1/auth/signup
// ──────────────────────────────────────────

type SignupRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// @Summary Register new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body SignupRequest true "Signup"
// @Success 201 {object} models.APIResponse
// @Router /auth/signup [post]
func (h *AuthHandler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp(err.Error()))
		return
	}

	// Check existing
	existing, _ := h.userRepo.FindByEmail(req.Email)
	if existing != nil {
		c.JSON(http.StatusConflict, models.ErrorResp("email already registered"))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), h.bcryptCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to hash password"))
		return
	}
	hashStr := string(hash)

	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: &hashStr,
		AuthProvider: models.AuthEmail,
	}

	if err := h.userRepo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to create user"))
		return
	}

	// Create default preferences
	prefs := &models.UserPreferences{UserID: user.ID}
	h.userRepo.CreatePreferences(prefs)

	tokens, err := h.jwtSvc.GenerateTokenPair(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to generate tokens"))
		return
	}

	setAuthCookies(c, tokens)
	c.JSON(http.StatusCreated, models.Success(gin.H{
		"user":   sanitizeUser(user),
		"tokens": tokens,
	}))
}

// ──────────────────────────────────────────
// POST /api/v1/auth/login
// ──────────────────────────────────────────

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp(err.Error()))
		return
	}

	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResp("invalid credentials"))
		return
	}

	if user.PasswordHash == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResp("please sign in with your OAuth provider"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResp("invalid credentials"))
		return
	}

	now := time.Now()
	user.LastLoginAt = &now
	h.userRepo.Update(user)

	tokens, err := h.jwtSvc.GenerateTokenPair(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to generate tokens"))
		return
	}

	setAuthCookies(c, tokens)
	c.JSON(http.StatusOK, models.Success(gin.H{
		"user":   sanitizeUser(user),
		"tokens": tokens,
	}))
}

// ──────────────────────────────────────────
// POST /api/v1/auth/refresh
// ──────────────────────────────────────────

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	// Try body first, then cookie
	if err := c.ShouldBindJSON(&req); err != nil {
		if cookie, cookieErr := c.Cookie("refresh_token"); cookieErr == nil {
			req.RefreshToken = cookie
		} else {
			c.JSON(http.StatusBadRequest, models.ErrorResp("refresh_token required"))
			return
		}
	}

	userID, err := h.jwtSvc.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResp("invalid or expired refresh token"))
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResp("user not found"))
		return
	}

	tokens, err := h.jwtSvc.GenerateTokenPair(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to generate tokens"))
		return
	}

	setAuthCookies(c, tokens)
	c.JSON(http.StatusOK, models.Success(tokens))
}

// ──────────────────────────────────────────
// GET /api/v1/auth/oauth/:provider
// ──────────────────────────────────────────

func (h *AuthHandler) OAuthRedirect(c *gin.Context) {
	provider := c.Param("provider")
	state := uuid.New().String()

	// Store state in cookie for verification
	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	url, err := h.oauthSvc.GetAuthURL(provider, state)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp("unsupported provider"))
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, url)
}

// ──────────────────────────────────────────
// GET /api/v1/auth/oauth/:provider/callback
// ──────────────────────────────────────────

func (h *AuthHandler) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")

	expectedState, err := c.Cookie("oauth_state")
	if err != nil || state != expectedState {
		c.JSON(http.StatusBadRequest, models.ErrorResp("invalid state parameter"))
		return
	}
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	userInfo, err := h.oauthSvc.ExchangeAndGetUser(c.Request.Context(), provider, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("oauth exchange failed"))
		return
	}

	// Find or create user
	user, err := h.userRepo.FindByEmail(userInfo.Email)
	if err != nil || user == nil {
		// Create new OAuth user
		avatarURL := userInfo.AvatarURL
		providerID := userInfo.ID
		user = &models.User{
			ID:             uuid.New(),
			Email:          userInfo.Email,
			Name:           userInfo.Name,
			AvatarURL:      &avatarURL,
			AuthProvider:   models.AuthProvider(provider),
			AuthProviderID: &providerID,
			EmailVerified:  true,
		}
		if err := h.userRepo.Create(user); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to create user"))
			return
		}
		prefs := &models.UserPreferences{UserID: user.ID}
		h.userRepo.CreatePreferences(prefs)
	}

	tokens, err := h.jwtSvc.GenerateTokenPair(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to generate tokens"))
		return
	}

	setAuthCookies(c, tokens)

	// Redirect to frontend with token info
	frontendURL := c.Query("redirect_uri")
	if frontendURL == "" {
		frontendURL = "/dashboard"
	}
	c.Redirect(http.StatusTemporaryRedirect, frontendURL+"?login=success")
}

// ──────────────────────────────────────────
// POST /api/v1/auth/logout
// ──────────────────────────────────────────

func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, models.SuccessMsg("logged out"))
}

// ──────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────

func setAuthCookies(c *gin.Context, tokens *auth.TokenPair) {
	c.SetCookie("access_token", tokens.AccessToken, 86400, "/", "", false, true)
	c.SetCookie("refresh_token", tokens.RefreshToken, 86400*30, "/", "", false, true)
}

func sanitizeUser(u *models.User) gin.H {
	return gin.H{
		"id":              u.ID,
		"email":           u.Email,
		"name":            u.Name,
		"avatar_url":      u.AvatarURL,
		"role":            u.Role,
		"email_verified":  u.EmailVerified,
		"auth_provider":   u.AuthProvider,
		"created_at":      u.CreatedAt,
	}
}
