package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  Role   `json:"role"`
	} `json:"user"`
}

// LoginHandler authenticates a user by email/password and issues a JWT.
func LoginHandler(db *gorm.DB, jwtSecret string, jwtExpiryMinutes int) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
			return
		}

		user, err := FindUserByEmail(db, req.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		if err := ComparePassword(user.PasswordHash, req.Password); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		token, err := GenerateToken(user, jwtSecret, jwtExpiryMinutes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		resp := loginResponse{Token: token}
		resp.User.ID = user.ID
		resp.User.Name = user.Name
		resp.User.Email = user.Email
		resp.User.Role = user.Role

		c.JSON(http.StatusOK, resp)
	}
}

// RegisterRoutes mounts auth-related public routes.
func RegisterRoutes(router *gin.Engine, db *gorm.DB, jwtSecret string, jwtExpiryMinutes int) {
	router.POST("/auth/login", LoginHandler(db, jwtSecret, jwtExpiryMinutes))
}