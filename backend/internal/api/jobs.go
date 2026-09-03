package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/remotescope/backend/internal/auth"
	"github.com/remotescope/backend/internal/models"
	"github.com/remotescope/backend/internal/repository"
)

type JobsHandler struct {
	jobRepo  *repository.JobRepository
}

func NewJobsHandler(jobRepo *repository.JobRepository) *JobsHandler {
	return &JobsHandler{jobRepo: jobRepo}
}

// ──────────────────────────────────────────
// GET /api/v1/jobs
// ──────────────────────────────────────────

// @Summary List jobs with filters
// @Tags jobs
// @Produce json
// @Param q query string false "Search query"
// @Param category query string false "Category slug"
// @Param experience query []string false "Experience levels"
// @Param employment_type query []string false "Employment types"
// @Param remote_type query string false "Remote type"
// @Param salary_min query int false "Min salary"
// @Param salary_max query int false "Max salary"
// @Param tech query []string false "Technologies"
// @Param posted_within query string false "5m|15m|1h|24h|7d"
// @Param page query int false "Page" default(1)
// @Param per_page query int false "Per page" default(20)
// @Success 200 {object} models.PaginatedResponse[models.Job]
// @Router /jobs [get]
func (h *JobsHandler) List(c *gin.Context) {
	var filters models.JobFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp(err.Error()))
		return
	}

	userID, _ := auth.GetUserID(c)
	var uid *uuid.UUID
	if userID != uuid.Nil {
		uid = &userID
	}

	jobs, total, err := h.jobRepo.List(&filters, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to fetch jobs"))
		return
	}

	perPage := filters.PerPage
	if perPage == 0 {
		perPage = 20
	}
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, models.Success(models.PaginatedResponse[models.Job]{
		Data:       jobs,
		Total:      total,
		Page:       filters.Page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}))
}

// ──────────────────────────────────────────
// GET /api/v1/jobs/:id
// ──────────────────────────────────────────

func (h *JobsHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp("invalid job ID"))
		return
	}

	job, err := h.jobRepo.FindByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResp("job not found"))
		return
	}

	// Record view asynchronously
	go h.jobRepo.IncrementView(id)

	c.JSON(http.StatusOK, models.Success(job))
}

// ──────────────────────────────────────────
// GET /api/v1/jobs/latest
// ──────────────────────────────────────────

func (h *JobsHandler) Latest(c *gin.Context) {
	limit := 20
	jobs, err := h.jobRepo.GetLatest(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to fetch latest jobs"))
		return
	}
	c.JSON(http.StatusOK, models.Success(jobs))
}

// ──────────────────────────────────────────
// GET /api/v1/jobs/trending
// ──────────────────────────────────────────

func (h *JobsHandler) Trending(c *gin.Context) {
	jobs, err := h.jobRepo.GetTrending(20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to fetch trending jobs"))
		return
	}
	c.JSON(http.StatusOK, models.Success(jobs))
}

// ──────────────────────────────────────────
// GET /api/v1/recommendations
// ──────────────────────────────────────────

func (h *JobsHandler) Recommendations(c *gin.Context) {
	userID, ok := auth.GetUserID(c)
	if !ok {
		// Guest: return latest
		jobs, _ := h.jobRepo.GetLatest(20)
		c.JSON(http.StatusOK, models.Success(jobs))
		return
	}

	jobs, err := h.jobRepo.GetRecommended(userID, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to fetch recommendations"))
		return
	}
	c.JSON(http.StatusOK, models.Success(jobs))
}

// ──────────────────────────────────────────
// POST /api/v1/jobs/:id/save
// ──────────────────────────────────────────

func (h *JobsHandler) SaveJob(jobRepo *repository.JobRepository, savedRepo *repository.SavedJobRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := auth.GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.ErrorResp("authentication required"))
			return
		}

		jobID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResp("invalid job ID"))
			return
		}

		if err := savedRepo.Save(userID, jobID); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to save job"))
			return
		}

		c.JSON(http.StatusOK, models.SuccessMsg("job saved"))
	}
}

func (h *JobsHandler) UnsaveJob(savedRepo *repository.SavedJobRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := auth.GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.ErrorResp("authentication required"))
			return
		}
		jobID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResp("invalid job ID"))
			return
		}
		if err := savedRepo.Unsave(userID, jobID); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to unsave job"))
			return
		}
		c.JSON(http.StatusOK, models.SuccessMsg("job unsaved"))
	}
}

func (h *JobsHandler) HideJob(savedRepo *repository.SavedJobRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := auth.GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.ErrorResp("authentication required"))
			return
		}
		jobID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResp("invalid job ID"))
			return
		}

		var body struct {
			Reason string `json:"reason"`
		}
		c.ShouldBindJSON(&body)

		if err := savedRepo.Hide(userID, jobID, body.Reason); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to hide job"))
			return
		}
		c.JSON(http.StatusOK, models.SuccessMsg("job hidden"))
	}
}

func (h *JobsHandler) ApplyJob(savedRepo *repository.SavedJobRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := auth.GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.ErrorResp("authentication required"))
			return
		}
		jobID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, models.ErrorResp("invalid job ID"))
			return
		}

		if err := savedRepo.MarkApplied(userID, jobID); err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to mark job as applied"))
			return
		}
		c.JSON(http.StatusOK, models.SuccessMsg("marked as applied"))
	}
}
