package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/remotescope/backend/internal/models"
	"github.com/remotescope/backend/pkg/config"
)

// ──────────────────────────────────────────
// JWT Claims
// ──────────────────────────────────────────

type Claims struct {
	UserID uuid.UUID        `json:"user_id"`
	Email  string           `json:"email"`
	Role   models.UserRole  `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// ──────────────────────────────────────────
// JWT Service
// ──────────────────────────────────────────

type JWTService struct {
	cfg *config.Config
}

func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{cfg: cfg}
}

func (s *JWTService) GenerateTokenPair(user *models.User) (*TokenPair, error) {
	// Access token
	accessExpiry := time.Now().Add(time.Duration(s.cfg.JWT.ExpiryHours) * time.Hour)
	accessClaims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessStr, err := accessToken.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	// Refresh token (longer-lived, minimal claims)
	refreshExpiry := time.Now().AddDate(0, 0, s.cfg.JWT.RefreshExpiryDays)
	refreshClaims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(refreshExpiry),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Subject:   user.ID.String(),
		ID:        uuid.New().String(), // jti for revocation
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshStr, err := refreshToken.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessStr,
		RefreshToken: refreshStr,
		ExpiresAt:    accessExpiry,
	}, nil
}

func (s *JWTService) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func (s *JWTService) ValidateRefreshToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return uuid.Nil, errors.New("invalid refresh token")
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID in token")
	}
	return userID, nil
}

// ──────────────────────────────────────────
// Middleware: Auth (required)
// ──────────────────────────────────────────

func RequireAuth(jwtSvc *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResp("authentication required"))
			return
		}
		claims, err := jwtSvc.ValidateAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResp("invalid or expired token"))
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// Middleware: Auth (optional — sets context if token present)
func OptionalAuth(jwtSvc *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token != "" {
			if claims, err := jwtSvc.ValidateAccessToken(token); err == nil {
				c.Set("user_id", claims.UserID)
				c.Set("user_email", claims.Email)
				c.Set("user_role", claims.Role)
			}
		}
		c.Next()
	}
}

// Middleware: Admin only
func RequireAdmin(jwtSvc *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResp("authentication required"))
			return
		}
		claims, err := jwtSvc.ValidateAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResp("invalid or expired token"))
			return
		}
		if claims.Role != models.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResp("admin access required"))
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

func extractBearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if len(header) > 7 && header[:7] == "Bearer " {
		return header[7:]
	}
	// Also check cookie
	cookie, err := c.Cookie("access_token")
	if err == nil {
		return cookie
	}
	return ""
}

// Helper: get user ID from context
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	v, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func GetUserRole(c *gin.Context) (models.UserRole, bool) {
	v, exists := c.Get("user_role")
	if !exists {
		return "", false
	}
	role, ok := v.(models.UserRole)
	return role, ok
}
