package team

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

type createTeamRequest struct {
	Name                string `json:"name" binding:"required"`
	LogoURL             string `json:"logo_url"`
	FoundedYear         int    `json:"founded_year" binding:"required"`
	HeadquartersAddress string `json:"headquarters_address"`
	HeadquartersCity    string `json:"headquarters_city"`
	ManagerUserID       *uint  `json:"manager_user_id"`
}

type updateTeamRequest struct {
	Name                string `json:"name"`
	LogoURL             string `json:"logo_url"`
	FoundedYear         int    `json:"founded_year"`
	HeadquartersAddress string `json:"headquarters_address"`
	HeadquartersCity    string `json:"headquarters_city"`
	ManagerUserID       *uint  `json:"manager_user_id"`
}

func (h *Handler) ListTeams(c *gin.Context) {
	teams, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch teams"})
		return
	}
	if teams == nil {
		teams = []Team{}
	}
	c.JSON(http.StatusOK, gin.H{"data": teams})
}

func (h *Handler) GetTeam(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}

	team, err := h.service.GetByID(uint(id))
	if err != nil {
		if err == ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch team"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": team})
}

func (h *Handler) CreateTeam(c *gin.Context) {
	role, _ := c.Get(auth.ContextUserRoleKey)
	userRole, ok := role.(auth.Role)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req createTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team := Team{
		Name:                req.Name,
		LogoURL:             req.LogoURL,
		FoundedYear:         req.FoundedYear,
		HeadquartersAddress: req.HeadquartersAddress,
		HeadquartersCity:    req.HeadquartersCity,
		ManagerUserID:       req.ManagerUserID,
	}

	if err := h.service.Create(&team, userRole); err != nil {
		if err == ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == ErrNameRequired {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create team"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": team})
}

func (h *Handler) UpdateTeam(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}

	userIDVal, _ := c.Get(auth.ContextUserIDKey)
	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	roleVal, _ := c.Get(auth.ContextUserRoleKey)
	userRole, ok := roleVal.(auth.Role)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req updateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := Team{
		Name:                req.Name,
		LogoURL:             req.LogoURL,
		FoundedYear:         req.FoundedYear,
		HeadquartersAddress: req.HeadquartersAddress,
		HeadquartersCity:    req.HeadquartersCity,
		ManagerUserID:       req.ManagerUserID,
	}
	updates.ID = uint(id)

	updated, err := h.service.Update(&updates, userID, userRole)
	if err != nil {
		if err == ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
			return
		}
		if err == ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == ErrNameRequired {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update team"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": updated})
}

func (h *Handler) DeleteTeam(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("team_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}

	roleVal, _ := c.Get(auth.ContextUserRoleKey)
	userRole, ok := roleVal.(auth.Role)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.service.Delete(uint(id), userRole); err != nil {
		if err == ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "team not found"})
			return
		}
		if err == ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete team"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "team deleted successfully"})
}

func RegisterRoutes(router *gin.Engine, db *gorm.DB, jwtSecret string) {
	repo := NewRepository(db)
	svc := NewService(repo)
	h := NewHandler(svc)

	authMW := auth.Middleware(jwtSecret)
	adminOnly := auth.RequireRole(auth.RoleAdmin)
	managerOrAdmin := auth.RequireRole(auth.RoleAdmin, auth.RoleManager)

	teams := router.Group("/api/teams")
	teams.Use(authMW)
	{
		teams.GET("", h.ListTeams)
		teams.GET("/:team_id", h.GetTeam)
		teams.POST("", adminOnly, h.CreateTeam)
		teams.PUT("/:team_id", managerOrAdmin, h.UpdateTeam)
		teams.DELETE("/:team_id", adminOnly, h.DeleteTeam)
	}
}
