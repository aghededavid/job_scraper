package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/remotescope/backend/internal/models"
	"github.com/remotescope/backend/internal/repository"
)

type CompaniesHandler struct {
	companyRepo *repository.CompanyRepository
}

func NewCompaniesHandler(companyRepo *repository.CompanyRepository) *CompaniesHandler {
	return &CompaniesHandler{companyRepo: companyRepo}
}

func (h *CompaniesHandler) List(c *gin.Context) {
	query := c.Query("q")
	page := 1
	perPage := 20

	companies, total, err := h.companyRepo.List(page, perPage, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResp("failed to fetch companies"))
		return
	}
	c.JSON(http.StatusOK, models.Success(gin.H{
		"data":  companies,
		"total": total,
	}))
}

func (h *CompaniesHandler) Get(c *gin.Context) {
	slug := c.Param("slug")
	company, err := h.companyRepo.FindBySlug(slug)
	if err != nil || company == nil {
		c.JSON(http.StatusNotFound, models.ErrorResp("company not found"))
		return
	}
	c.JSON(http.StatusOK, models.Success(company))
}
