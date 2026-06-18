package report

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"kickoff/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetReport(c *gin.Context) {
	matchID, err := strconv.ParseUint(c.Param("match_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}

	report, err := h.service.GetReport(uint(matchID))
	if err != nil {
		switch err {
		case ErrMatchNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case ErrMatchNotFinished:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": report})
}

func RegisterRoutes(router *gin.Engine, db *gorm.DB, jwtSecret string) {
	svc := NewService(db)
	h := NewHandler(svc)

	router.GET("/api/matches/:match_id/report", auth.Middleware(jwtSecret), h.GetReport)
}
