package app

import (
	"time"

	"sermon-backend/internal/config"
	"sermon-backend/internal/handlers"
	"sermon-backend/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	authHandler := handlers.NewAuthHandler(db, cfg)
	serverHandler := handlers.NewServerHandler(db)
	agentHandler := handlers.NewAgentHandler(db)
	incidentHandler := handlers.NewIncidentHandler(db)
	thresholdHandler := handlers.NewThresholdHandler(db)
	userHandler := handlers.NewUserHandler(db)

	api := r.Group("/api/v1")
	{
		api.POST("/auth/login", authHandler.Login)
		api.POST("/agent/metrics", agentHandler.ReceiveMetrics)

		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			protected.GET("/auth/me", authHandler.Me)

			protected.GET("/servers", serverHandler.GetServers)
			protected.GET("/servers/:id", serverHandler.GetServerByID)
			protected.GET("/servers/:id/metrics/current", serverHandler.GetCurrentMetrics)
			protected.GET("/servers/:id/metrics/history", serverHandler.GetMetricHistory)

			protected.GET("/incidents", incidentHandler.GetIncidents)
			protected.GET("/incidents/:id", incidentHandler.GetIncidentByID)
			protected.PATCH("/incidents/:id/status", incidentHandler.UpdateIncidentStatus)
			protected.GET("/incidents/:id/comments", incidentHandler.GetComments)
			protected.POST("/incidents/:id/comments", incidentHandler.AddComment)

			protected.GET("/servers/:id/thresholds", thresholdHandler.GetServerThresholds)
			protected.PUT("/servers/:id/thresholds/:metric_type", middleware.RequireRole("admin"), thresholdHandler.UpdateServerThreshold)
			protected.POST("/auth/change-password", authHandler.ChangePassword)
			protected.GET("/users", middleware.RequireRole("admin"), userHandler.GetUsers)
			protected.POST("/users", middleware.RequireRole("admin"), userHandler.CreateUser)
			protected.POST("/users/:id/reset-password", middleware.RequireRole("admin"), userHandler.ResetPassword)
			protected.PATCH("/users/:id/deactivate", middleware.RequireRole("admin"), userHandler.DeactivateUser)
			protected.PATCH("/users/:id/activate", middleware.RequireRole("admin"), userHandler.ActivateUser)
			protected.POST("/servers", middleware.RequireRole("admin"), serverHandler.CreateServer)
			protected.PATCH("/servers/:id", middleware.RequireRole("admin"), serverHandler.UpdateServer)
			protected.PATCH("/servers/:id/deactivate", middleware.RequireRole("admin"), serverHandler.DeactivateServer)
			protected.PATCH("/servers/:id/activate", middleware.RequireRole("admin"), serverHandler.ActivateServer)
			protected.POST("/servers/:id/regenerate-agent-token", middleware.RequireRole("admin"), serverHandler.RegenerateAgentToken)

		}
	}

	return r
}
