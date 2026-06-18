package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"kickoff/internal/auth"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecureP@ss123"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("hash is empty")
	}

	if err := auth.ComparePassword(hash, password); err != nil {
		t.Fatal("ComparePassword should match")
	}
	if err := auth.ComparePassword(hash, "wrong"); err == nil {
		t.Fatal("ComparePassword should reject wrong password")
	}
}

func TestJWTTokens(t *testing.T) {
	secret := "test-secret"
	user := &auth.User{ID: 1, Email: "a@b.com", Role: auth.RoleAdmin}

	token, err := auth.GenerateToken(user, secret, 60)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := auth.ParseToken(token, secret)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if claims.UserID != 1 {
		t.Fatalf("expected UserID 1, got %d", claims.UserID)
	}
	if claims.Email != "a@b.com" {
		t.Fatalf("expected email a@b.com, got %s", claims.Email)
	}
	if claims.Role != auth.RoleAdmin {
		t.Fatalf("expected role ADMIN, got %s", claims.Role)
	}
}

func TestJWTExpired(t *testing.T) {
	secret := "test-secret"
	user := &auth.User{ID: 1, Email: "a@b.com", Role: auth.RoleUser}

	token, err := auth.GenerateToken(user, secret, -1)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if _, err := auth.ParseToken(token, secret); err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestJWTWrongSecret(t *testing.T) {
	user := &auth.User{ID: 1, Email: "a@b.com", Role: auth.RoleUser}
	token, _ := auth.GenerateToken(user, "real-secret", 60)
	if _, err := auth.ParseToken(token, "wrong-secret"); err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestJWTBadToken(t *testing.T) {
	if _, err := auth.ParseToken("invalid.jwt.string", "secret"); err == nil {
		t.Fatal("expected error for malformed token")
	}
}

func TestMiddlewareNoHeader(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	auth.Middleware("secret")(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestMiddlewareBadHeader(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "NotBearer token")

	auth.Middleware("secret")(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestMiddlewareValidToken(t *testing.T) {
	secret := "test-secret"
	user := &auth.User{ID: 1, Email: "a@b.com", Role: auth.RoleAdmin}
	token, _ := auth.GenerateToken(user, secret, 60)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	auth.Middleware(secret)(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	uid, _ := c.Get(auth.ContextUserIDKey)
	if uid != user.ID {
		t.Fatalf("expected userID %d, got %d", user.ID, uid)
	}
	role, _ := c.Get(auth.ContextUserRoleKey)
	if role != user.Role {
		t.Fatalf("expected role %s, got %s", user.Role, role)
	}
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name     string
		userRole auth.Role
		allowed  []auth.Role
		wantCode int
	}{
		{"admin allowed", auth.RoleAdmin, []auth.Role{auth.RoleAdmin}, http.StatusOK},
		{"manager denied admin", auth.RoleManager, []auth.Role{auth.RoleAdmin}, http.StatusForbidden},
		{"user denied admin", auth.RoleUser, []auth.Role{auth.RoleAdmin}, http.StatusForbidden},
		{"manager allowed manager", auth.RoleManager, []auth.Role{auth.RoleManager}, http.StatusOK},
		{"user allowed user", auth.RoleUser, []auth.Role{auth.RoleUser}, http.StatusOK},
		{"admin allowed multi", auth.RoleAdmin, []auth.Role{auth.RoleAdmin, auth.RoleManager}, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set(auth.ContextUserRoleKey, tt.userRole)

			auth.RequireRole(tt.allowed...)(c)
			if w.Code != tt.wantCode {
				t.Fatalf("expected code %d, got %d", tt.wantCode, w.Code)
			}
		})
	}
}
