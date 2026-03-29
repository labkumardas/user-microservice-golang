package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"user-microservice-golang/config"
	"user-microservice-golang/controller"
	"user-microservice-golang/middleware"
	"user-microservice-golang/model"
)

// Setup registers all routes and returns the configured Gin engine
func Setup(
	cfg *config.AppConfig,
	ctrl *controller.UserController,
	logger *zap.Logger,
) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// ── Global middleware ────────────────────────────────────────────────────
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.RateLimit(100, time.Minute)) // 100 req/min per IP

	// CORS headers (extend as needed for your gateway)
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// ── System ──────────────────────────────────────────────────────────────
	r.GET("/health", ctrl.HealthCheck)

	// ── API v1 ──────────────────────────────────────────────────────────────
	v1 := r.Group("/api/v1")
	{
		// Public – no auth required
		auth := v1.Group("/auth")
		{
			auth.POST("/register", ctrl.Register)
			auth.POST("/login", ctrl.Login)
		}

		// Protected – valid JWT required
		users := v1.Group("/users")
		users.Use(middleware.Authenticate(cfg.JWTSecret, logger))
		{
			// Any authenticated user
			users.GET("/me", ctrl.GetMe)

			// Self or admin
			users.GET("/:id", middleware.SelfOrAdmin(), ctrl.GetUserByID)
			users.PATCH("/:id", middleware.SelfOrAdmin(), ctrl.UpdateProfile)
			users.PATCH("/:id/password", middleware.SelfOrAdmin(), ctrl.UpdatePassword)
			users.DELETE("/:id", middleware.SelfOrAdmin(), ctrl.DeleteUser)

			// Admin only
			users.GET("", middleware.RequireRole(string(model.RoleAdmin)), ctrl.GetAllUsers)
			users.PATCH("/:id/status", middleware.RequireRole(string(model.RoleAdmin)), ctrl.UpdateStatus)
		}
	}

	return r
}
