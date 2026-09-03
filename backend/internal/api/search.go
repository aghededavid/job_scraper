package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/remotescope/backend/internal/models"
	"github.com/remotescope/backend/internal/repository"
	"github.com/remotescope/backend/pkg/config"
)

type SearchHandler struct {
	jobRepo *repository.JobRepository
	cfg     *config.Config
}

func NewSearchHandler(jobRepo *repository.JobRepository, cfg *config.Config) *SearchHandler {
	return &SearchHandler{jobRepo: jobRepo, cfg: cfg}
}

type SearchRequest struct {
	Query   string            `json:"query"`
	Filters models.JobFilters `json:"filters"`
	Mode    string            `json:"mode"` // "hybrid" | "semantic" | "keyword"
}

// POST /api/v1/search
// @Summary Natural language + filter search
// @Tags search
// @Accept json
// @Produce json
// @Param body body SearchRequest true "Search request"
// @Success 200 {object} models.PaginatedResponse[models.Job]
// @Router /search [post]
func (h *SearchHandler) Search(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResp(err.Error()))
		return
	}

	// Merge top-level query into filters
	if req.Query != "" {
		req.Filters.Query = req.Query
	}

	mode := req.Mode
	if mode == "" {
		mode = "hybrid"
	}

	var jobs []models.Job
	var total int64
	var err error

	switch mode {
	case "semantic":
		// Get embedding from AI service
		embedding, embErr := h.getEmbedding(c.Request.Context(), req.Query)
		if embErr == nil && len(embedding) > 0 {
			jobs, err = h.jobRepo.VectorSearch(embedding, req.Filters.PerPage, &req.Filters)
			total = int64(len(jobs))
		} else {
			// Fallback to keyword
			jobs, total, err = h.jobRepo.List(&req.Filters, nil)
		}
	case "hybrid":
		// Vector search for ranking + keyword for recall
		embedding, embErr := h.getEmbedding(c.Request.Context(), req.Query)
		if embErr == nil && len(embedding) > 0 {
			vectorJobs, vErr := h.jobRepo.VectorSearch(embedding, 50, &req.Filters)
			if vErr == nil && len(vectorJobs) > 0 {
				jobs = vectorJobs
				total = int64(len(vectorJobs))
				break
			}
		}
		fallthrough
	default:
		jobs, total, err = h.jobRepo.List(&req.Filters, nil)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("search failed"))
		return
	}

	perPage := req.Filters.PerPage
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
		Page:       req.Filters.Page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}))
}

func (h *SearchHandler) getEmbedding(ctx context.Context, text string) ([]float32, error) {
	if text == "" || h.cfg.AI.ServiceURL == "" {
		return nil, fmt.Errorf("no query or AI service")
	}

	payload, _ := json.Marshal(map[string]string{"text": text})
	req, err := http.NewRequestWithContext(ctx, "POST",
		h.cfg.AI.ServiceURL+"/embed",
		bytesReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Embedding, nil
}

// Suggestions autocomplete
// GET /api/v1/search/suggest?q=...
func (h *SearchHandler) Suggest(c *gin.Context) {
	q := c.Query("q")
	if len(q) < 2 {
		c.JSON(http.StatusOK, models.Success([]string{}))
		return
	}

	var titles []string
	h.jobRepo.DB().Raw(`
		SELECT DISTINCT normalized_title 
		FROM jobs 
		WHERE normalized_title ILIKE ? AND is_active = true
		LIMIT 10
	`, q+"%").Scan(&titles)

	c.JSON(http.StatusOK, models.Success(titles))
}

// bytesReader creates io.Reader from bytes
func bytesReader(b []byte) *bytesReaderImpl {
	return &bytesReaderImpl{data: b, pos: 0}
}

type bytesReaderImpl struct {
	data []byte
	pos  int
}

func (r *bytesReaderImpl) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, fmt.Errorf("EOF")
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
