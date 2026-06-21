package main

import (
	"log"

	"config-center/internal/config"
	"config-center/internal/database"
	"config-center/internal/handlers"
	"config-center/internal/push"
	redisclient "config-center/internal/redisclient"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if err := database.Init(cfg); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	if err := redisclient.Init(cfg); err != nil {
		log.Fatalf("Failed to init redis: %v", err)
	}

	push.Init(cfg)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	configHandler := handlers.NewConfigHandler()
	namespaceHandler := handlers.NewNamespaceHandler()
	pushHandler := handlers.NewPushHandler()
	grayHandler := handlers.NewGrayHandler()
	wsHandler := handlers.NewWSHandler()
	metricHandler := handlers.NewMetricHandler()

	api := r.Group("/api/v1")
	{
		namespaces := api.Group("/namespaces")
		{
			namespaces.GET("", namespaceHandler.ListNamespaces)
			namespaces.POST("", namespaceHandler.CreateNamespace)
			namespaces.GET("/:id", namespaceHandler.GetNamespace)
			namespaces.PUT("/:id", namespaceHandler.UpdateNamespace)
			namespaces.DELETE("/:id", namespaceHandler.DeleteNamespace)
		}

		groups := api.Group("/groups")
		{
			groups.GET("", namespaceHandler.ListGroups)
			groups.POST("", namespaceHandler.CreateGroup)
			groups.GET("/:id", namespaceHandler.GetGroup)
			groups.PUT("/:id", namespaceHandler.UpdateGroup)
			groups.DELETE("/:id", namespaceHandler.DeleteGroup)
		}

		configs := api.Group("/configs")
		{
			configs.GET("", configHandler.ListConfigItems)
			configs.POST("", configHandler.CreateConfigItem)
			configs.GET("/:id", configHandler.GetConfigItem)
			configs.PUT("/:id", configHandler.UpdateConfigItem)
			configs.DELETE("/:id", configHandler.DeleteConfigItem)
			configs.POST("/validate", configHandler.ValidateConfig)
			configs.GET("/merged", configHandler.GetMergedConfig)
			configs.POST("/:id/rollback", configHandler.RollbackVersion)
			configs.GET("/:id/versions", configHandler.GetVersionHistory)
			configs.GET("/:id/compare", configHandler.CompareVersions)
			configs.POST("/batch-delete", configHandler.BatchDeleteConfigItems)
			configs.POST("/batch-copy", configHandler.BatchCopyConfigItems)
		}

		gray := api.Group("/gray")
		{
			gray.GET("", grayHandler.ListGrayReleases)
			gray.POST("", grayHandler.CreateGrayRelease)
			gray.GET("/:id", grayHandler.GetGrayRelease)
			gray.POST("/:id/start", grayHandler.StartGrayRelease)
			gray.POST("/:id/full-push", grayHandler.FullPush)
			gray.POST("/:id/rollback", grayHandler.RollbackGrayRelease)
		}

		push := api.Group("/push")
		{
			push.GET("/long-poll", pushHandler.LongPoll)
			push.GET("/connections", pushHandler.GetConnections)
			push.GET("/stats", pushHandler.GetConnectionStats)
		}

		ws := api.Group("/ws")
		{
			ws.GET("", wsHandler.HandleWebSocket)
		}

		metrics := api.Group("/metrics")
		{
			metrics.GET("", metricHandler.GetMetrics)
			metrics.GET("/latest", metricHandler.GetLatestMetrics)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	log.Printf("Server starting on port %d", cfg.ServerPort)
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
