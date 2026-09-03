package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/remotescope/backend/internal/auth"
	"github.com/remotescope/backend/internal/models"
	"github.com/remotescope/backend/internal/repository"
)

type UsersHandler struct {
	userRepo *repository.UserRepository
}

func NewUsersHandler(userRepo *repository.UserRepository) *UsersHandler {
	return &UsersHandler{userRepo: userRepo}
}

// GET /api/v1/users/me
func (h *UsersHandler) GetMe(c *gin.Context) {
	userID, _ := auth.GetUserID(c)
	user, err := h.userRepo.FindByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, models.ErrorResp("user not found"))
		return
	}
	c.JSON(http.StatusOK, models.Success(user))
}

// PUT /api/v1/users/me
type UpdateProfileRequest struct {
	Name      string `json:"name" binding:"omitempty,min=2,max=100"`
	AvatarURL string `json:"avatar_url"`
}

func (h *UsersHandler) UpdateMe(c *gin.Context) {
	userID, _ := auth.GetUserID(c)
	user, err := h.userRepo.FindByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, models.ErrorResp("user not found"))
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp(err.Error()))
		return
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.AvatarURL != "" {
		user.AvatarURL = &req.AvatarURL
	}

	if err := h.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to update profile"))
		return
	}
	c.JSON(http.StatusOK, models.Success(user))
}

// PUT /api/v1/users/me/preferences
func (h *UsersHandler) UpdatePreferences(c *gin.Context) {
	userID, _ := auth.GetUserID(c)

	var prefs models.UserPreferences
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp(err.Error()))
		return
	}
	prefs.UserID = userID

	existing, _ := h.userRepo.GetPreferences(userID)
	if existing == nil {
		prefs.ID = [16]byte{}
		if err := h.userRepo.CreatePreferences(&prefs); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to create preferences"))
			return
		}
	} else {
		prefs.ID = existing.ID
		if err := h.userRepo.UpdatePreferences(&prefs); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to update preferences"))
			return
		}
	}

	c.JSON(http.StatusOK, models.Success(&prefs))
}

// POST /api/v1/users/me/complete-onboarding
func (h *UsersHandler) CompleteOnboarding(c *gin.Context) {
	userID, _ := auth.GetUserID(c)
	prefs, _ := h.userRepo.GetPreferences(userID)
	if prefs == nil {
		newPrefs := &models.UserPreferences{UserID: userID, OnboardingCompleted: true}
		h.userRepo.CreatePreferences(newPrefs)
	} else {
		prefs.OnboardingCompleted = true
		h.userRepo.UpdatePreferences(prefs)
	}
	c.JSON(http.StatusOK, models.SuccessMsg("onboarding completed"))
}
