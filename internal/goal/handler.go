package goal

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

type createGoalRequest struct {
	PlayerID   uint `json:"player_id" binding:"required"`
	GoalMinute int  `json:"goal_minute" binding:"required"`
}

func (h *Handler) ListByMatch(c *gin.Context) {
	matchID, err := strconv.ParseUint(c.Param("match_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}

	goals, err := h.service.ListByMatch(uint(matchID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch goals"})
		return
	}
	if goals == nil {
		goals = []Goal{}
	}
	c.JSON(http.StatusOK, gin.H{"data": goals})
}

func (h *Handler) Create(c *gin.Context) {
	matchID, err := strconv.ParseUint(c.Param("match_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}

	roleVal, _ := c.Get(auth.ContextUserRoleKey)
	role, ok := roleVal.(auth.Role)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req createGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	goal := Goal{
		MatchID:    uint(matchID),
		PlayerID:   req.PlayerID,
		GoalMinute: req.GoalMinute,
	}

	if err := h.service.Create(&goal, role); err != nil {
		switch err {
		case ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case ErrMatchNotFound, ErrPlayerNotFound, ErrPlayerNotInMatch, ErrInvalidMinute:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to record goal"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": goal})
}

func (h *Handler) Delete(c *gin.Context) {
	matchID, err := strconv.ParseUint(c.Param("match_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}

	goalID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal id"})
		return
	}

	roleVal, _ := c.Get(auth.ContextUserRoleKey)
	role, ok := roleVal.(auth.Role)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.service.Delete(uint(goalID), role); err != nil {
		switch err {
		case ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete goal"})
		}
		return
	}

	if err != nil {
		_ = matchID // matchID validated but unused by service
	}
	c.JSON(http.StatusOK, gin.H{"message": "goal deleted successfully"})
}

func RegisterRoutes(router *gin.Engine, db *gorm.DB, jwtSecret string) {
	repo := NewRepository(db)
	svc := NewService(repo, db)
	h := NewHandler(svc)

	authMW := auth.Middleware(jwtSecret)
	adminOnly := auth.RequireRole(auth.RoleAdmin)

	goals := router.Group("/api/matches/:match_id/goals")
	goals.Use(authMW)
	{
		goals.GET("", h.ListByMatch)
		goals.POST("", adminOnly, h.Create)
		goals.DELETE("/:id", adminOnly, h.Delete)
	}
}
