package player

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

type createPlayerRequest struct {
	Name         string   `json:"name" binding:"required"`
	HeightCm     float64  `json:"height_cm"`
	WeightKg     float64  `json:"weight_kg"`
	Position     Position `json:"position" binding:"required"`
	JerseyNumber int      `json:"jersey_number" binding:"required"`
}

type updatePlayerRequest struct {
	Name         string   `json:"name"`
	HeightCm     float64  `json:"height_cm"`
	WeightKg     float64  `json:"weight_kg"`
	Position     Position `json:"position"`
	JerseyNumber int      `json:"jersey_number"`
}

func (h *Handler) ListByTeam(c *gin.Context) {
	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}

	players, err := h.service.ListByTeam(uint(teamID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch players"})
		return
	}
	if players == nil {
		players = []Player{}
	}
	c.JSON(http.StatusOK, gin.H{"data": players})
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}

	player, err := h.service.GetByID(uint(id))
	if err != nil {
		if err == ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch player"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": player})
}

func (h *Handler) Create(c *gin.Context) {
	teamID, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}

	userID, _ := c.Get(auth.ContextUserIDKey)
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	roleVal, _ := c.Get(auth.ContextUserRoleKey)
	role, ok := roleVal.(auth.Role)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req createPlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	player := Player{
		TeamID:       uint(teamID),
		Name:         req.Name,
		HeightCm:     req.HeightCm,
		WeightKg:     req.WeightKg,
		Position:     req.Position,
		JerseyNumber: req.JerseyNumber,
	}

	if err := h.service.Create(&player, uid, role); err != nil {
		switch err {
		case ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case ErrTeamNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case ErrNameRequired, ErrInvalidPosition, ErrJerseyNumberRequired, ErrJerseyNumberTaken:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create player"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": player})
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}

	userID, _ := c.Get(auth.ContextUserIDKey)
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	roleVal, _ := c.Get(auth.ContextUserRoleKey)
	role, ok := roleVal.(auth.Role)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req updatePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := Player{
		Name:         req.Name,
		HeightCm:     req.HeightCm,
		WeightKg:     req.WeightKg,
		Position:     req.Position,
		JerseyNumber: req.JerseyNumber,
	}
	updates.ID = uint(id)

	updated, err := h.service.Update(&updates, uid, role)
	if err != nil {
		switch err {
		case ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case ErrNameRequired, ErrInvalidPosition, ErrJerseyNumberTaken:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update player"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updated})
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player id"})
		return
	}

	userID, _ := c.Get(auth.ContextUserIDKey)
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}
	roleVal, _ := c.Get(auth.ContextUserRoleKey)
	role, ok := roleVal.(auth.Role)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.service.Delete(uint(id), uid, role); err != nil {
		switch err {
		case ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case ErrForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete player"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "player deleted successfully"})
}

func RegisterRoutes(router *gin.Engine, db *gorm.DB, jwtSecret string) {
	repo := NewRepository(db)
	svc := NewService(repo, db)
	h := NewHandler(svc)

	authMW := auth.Middleware(jwtSecret)
	managerOrAdmin := auth.RequireRole(auth.RoleAdmin, auth.RoleManager)

	teams := router.Group("/api/teams")
	teams.Use(authMW)
	{
		teams.GET("/:team_id/players", h.ListByTeam)
		teams.POST("/:team_id/players", managerOrAdmin, h.Create)
	}

	players := router.Group("/api/players")
	players.Use(authMW)
	{
		players.GET("/:id", h.GetByID)
		players.PUT("/:id", managerOrAdmin, h.Update)
		players.DELETE("/:id", managerOrAdmin, h.Delete)
	}
}
