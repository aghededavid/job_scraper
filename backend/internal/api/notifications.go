package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/remotescope/backend/internal/auth"
	"github.com/remotescope/backend/internal/models"
	"github.com/remotescope/backend/internal/repository"
)

type NotificationsHandler struct {
	notifRepo *repository.NotificationRepository
}

func NewNotificationsHandler(notifRepo *repository.NotificationRepository) *NotificationsHandler {
	return &NotificationsHandler{notifRepo: notifRepo}
}

// GET /api/v1/notifications
func (h *NotificationsHandler) List(c *gin.Context) {
	userID, _ := auth.GetUserID(c)
	notifs, total, err := h.notifRepo.List(userID, 1, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to fetch notifications"))
		return
	}
	unread, _ := h.notifRepo.UnreadCount(userID)
	c.JSON(http.StatusOK, models.Success(gin.H{
		"items":  notifs,
		"total":  total,
		"unread": unread,
	}))
}

// PUT /api/v1/notifications/:id/read
func (h *NotificationsHandler) MarkRead(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp("invalid notification ID"))
		return
	}
	userID, _ := auth.GetUserID(c)
	if err := h.notifRepo.MarkRead(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to mark notification as read"))
		return
	}
	c.JSON(http.StatusOK, models.SuccessMsg("marked as read"))
}

// PUT /api/v1/notifications/read-all
func (h *NotificationsHandler) MarkAllRead(c *gin.Context) {
	userID, _ := auth.GetUserID(c)
	if err := h.notifRepo.MarkAllRead(userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to mark all as read"))
		return
	}
	c.JSON(http.StatusOK, models.SuccessMsg("all notifications marked as read"))
}

// GET /api/v1/saved-searches
func (h *NotificationsHandler) ListSavedSearches(c *gin.Context) {
	userID, _ := auth.GetUserID(c)
	searches, err := h.notifRepo.ListSavedSearches(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to fetch saved searches"))
		return
	}
	c.JSON(http.StatusOK, models.Success(searches))
}

type CreateSavedSearchRequest struct {
	Name    string      `json:"name" binding:"required"`
	Query   string      `json:"query"`
	Filters interface{} `json:"filters"`
	Notify  bool        `json:"notify"`
}

// POST /api/v1/saved-searches
func (h *NotificationsHandler) CreateSavedSearch(c *gin.Context) {
	userID, _ := auth.GetUserID(c)
	var req CreateSavedSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp(err.Error()))
		return
	}

	search := &models.SavedSearch{
		ID:      uuid.New(),
		UserID:  userID,
		Name:    req.Name,
		Filters: req.Filters,
		Notify:  req.Notify,
	}
	if req.Query != "" {
		search.Query = &req.Query
	}

	if err := h.notifRepo.CreateSavedSearch(search); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to create saved search"))
		return
	}
	c.JSON(http.StatusCreated, models.Success(search))
}

// DELETE /api/v1/saved-searches/:id
func (h *NotificationsHandler) DeleteSavedSearch(c *gin.Context) {
	userID, _ := auth.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp("invalid ID"))
		return
	}
	if err := h.notifRepo.DeleteSavedSearch(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to delete saved search"))
		return
	}
	c.JSON(http.StatusOK, models.SuccessMsg("deleted"))
}
