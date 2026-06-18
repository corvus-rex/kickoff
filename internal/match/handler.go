package match

import (
	"net/http"
	"strconv"
	"time"

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

type createMatchRequest struct {
	MatchDate  string      `json:"match_date" binding:"required"`
	MatchTime  string      `json:"match_time" binding:"required"`
	HomeTeamID uint        `json:"home_team_id" binding:"required"`
	AwayTeamID uint        `json:"away_team_id" binding:"required"`
	Status     MatchStatus `json:"status"`
}

type updateMatchRequest struct {
	MatchDate  string      `json:"match_date"`
	MatchTime  string      `json:"match_time"`
	HomeTeamID uint        `json:"home_team_id"`
	AwayTeamID uint        `json:"away_team_id"`
	Status     MatchStatus `json:"status"`
}

func (h *Handler) List(c *gin.Context) {
	matches, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch matches"})
		return
	}
	if matches == nil {
		matches = []Match{}
	}
	c.JSON(http.StatusOK, gin.H{"data": matches})
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("match_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match id"})
		return
	}

	m, err := h.service.GetByID(uint(id))
	if err != nil {
		if err == ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch match"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": m})
}

func (h *Handler) Create(c *gin.Context) {
	roleVal, _ := c.Get(auth.ContextUserRoleKey)
	role, ok := roleVal.(auth.Role)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req createMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	matchDate, err := time.Parse("2006-01-02", req.MatchDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match_date, expected YYYY-MM-DD"})
		return
	}

	if req.Status == "" {
		req.Status = StatusScheduled
	}

	m := Match{
		MatchDate:  matchDate,
		MatchTime:  req.MatchTime,
		HomeTeamID: req.HomeTeamID,
		AwayTeamID: req.AwayTeamID,
		Status:     req.Status,
	}

	if err := h.service.Create(&m, role); err != nil {
		switch err {
		case ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case ErrTeamNotFound, ErrSameTeam:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case ErrInvalidStatus, ErrInvalidDate, ErrInvalidTime, ErrDateRequired:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create match"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": m})
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("match_id"), 10, 64)
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

	var req updateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := Match{ID: uint(id)}
	if req.MatchDate != "" {
		parsed, err := time.Parse("2006-01-02", req.MatchDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid match_date, expected YYYY-MM-DD"})
			return
		}
		updates.MatchDate = parsed
	}
	if req.MatchTime != "" {
		updates.MatchTime = req.MatchTime
	}
	if req.HomeTeamID > 0 {
		updates.HomeTeamID = req.HomeTeamID
	}
	if req.AwayTeamID > 0 {
		updates.AwayTeamID = req.AwayTeamID
	}
	if req.Status != "" {
		updates.Status = req.Status
	}

	updated, err := h.service.Update(&updates, role)
	if err != nil {
		switch err {
		case ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case ErrSameTeam, ErrTeamNotFound, ErrInvalidStatus, ErrInvalidTime:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update match"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updated})
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("match_id"), 10, 64)
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

	if err := h.service.Delete(uint(id), role); err != nil {
		switch err {
		case ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete match"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "match deleted successfully"})
}

func (h *Handler) FinishMatch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("match_id"), 10, 64)
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

	m, err := h.service.Finish(uint(id), role)
	if err != nil {
		switch err {
		case ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case ErrAlreadyFinished:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to finish match"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": m})
}

func RegisterRoutes(router *gin.Engine, db *gorm.DB, jwtSecret string) {
	repo := NewRepository(db)
	svc := NewService(repo, db)
	h := NewHandler(svc)

	authMW := auth.Middleware(jwtSecret)
	adminOnly := auth.RequireRole(auth.RoleAdmin)

	matches := router.Group("/api/matches")
	matches.Use(authMW)
	{
		matches.GET("", h.List)
		matches.GET("/:match_id", h.GetByID)
		matches.POST("", adminOnly, h.Create)
		matches.PUT("/:match_id", adminOnly, h.Update)
		matches.PUT("/:match_id/finish", adminOnly, h.FinishMatch)
		matches.DELETE("/:match_id", adminOnly, h.Delete)
	}
}
