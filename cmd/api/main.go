package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"kickoff/internal/auth"
	"kickoff/internal/config"
	"kickoff/internal/database"
	"kickoff/internal/goal"
	"kickoff/internal/match"
	"kickoff/internal/player"
	"kickoff/internal/seed"
	"kickoff/internal/team"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on system environment variables")
	}

	cfg := config.Load()

	if cfg.Env != "development" && cfg.JWTSecret == config.DefaultJWTSecret {
		log.Fatal("JWT_SECRET must be explicitly set outside development environment")
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	database.RegisterModel(&auth.User{}, &team.Team{}, &player.Player{}, &match.Match{}, &goal.Goal{})

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	if cfg.SeedUsers {
		if err := auth.Seed(db); err != nil {
			log.Fatalf("user seeding failed: %v", err)
		}
	}

	if cfg.SeedDomain {
		if err := seed.Seed(db); err != nil {
			log.Fatalf("domain seeding failed: %v", err)
		}
	}

	router := setupRouter(db, cfg)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("starting server on port %s (env=%s)", cfg.Port, cfg.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed to start: %v", err)
		}
	}()

	gracefulShutdown(srv, db)
}

func setupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	router.GET("/health", healthCheckHandler)
	auth.RegisterRoutes(router, db, cfg.JWTSecret, cfg.JWTExpiryMinutes)
	team.RegisterRoutes(router, db, cfg.JWTSecret)
	player.RegisterRoutes(router, db, cfg.JWTSecret)
	match.RegisterRoutes(router, db, cfg.JWTSecret)
	goal.RegisterRoutes(router, db, cfg.JWTSecret)

	return router
}

func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func gracefulShutdown(srv *http.Server, db *gorm.DB) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutdown signal received, shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	if err := database.Close(db); err != nil {
		log.Printf("error closing database connection: %v", err)
	}

	log.Println("server exited cleanly")
}