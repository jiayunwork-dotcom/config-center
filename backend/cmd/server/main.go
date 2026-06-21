package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"config-center/internal/config"
	"config-center/internal/database"
	"config-center/internal/handlers"
	"config-center/internal/middleware"
	"config-center/internal/models"
	"config-center/internal/push"
	redisclient "config-center/internal/redisclient"
	"config-center/internal/services"

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
	authHandler := handlers.NewAuthHandler()
	auditHandler := handlers.NewAuditHandler()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.AuthMiddleware(), authHandler.Me)
		}

		admin := api.Group("")
		admin.Use(middleware.AuthMiddleware(), requireGlobalAdmin())
		{
			users := admin.Group("/users")
			{
				users.GET("", authHandler.ListUsers)
				users.POST("", authHandler.CreateUser)
				users.DELETE("/:id", authHandler.DeleteUser)
				users.GET("/:user_id/roles", authHandler.GetRoles)
			}
			admin.POST("/roles/grant", authHandler.GrantRole)
			admin.DELETE("/roles/:id", authHandler.RevokeRole)
		}

		authenticated := api.Group("")
		authenticated.Use(middleware.AuthMiddleware(), middleware.FilterNamespacesMiddleware())
		{
			audit := authenticated.Group("/audit-logs")
			{
				audit.GET("", requireGlobalAdminOrSelf(), auditHandler.ListLogs)
			}

			namespaces := authenticated.Group("/namespaces")
			{
				namespaces.GET("", namespaceHandler.ListNamespaces)
				namespaces.POST("", requireGlobalAdmin(), namespaceHandler.CreateNamespace)
				namespaces.GET("/:id", namespaceViewerRBAC(), namespaceHandler.GetNamespace)
				namespaces.PUT("/:id", requireGlobalAdmin(), namespaceHandler.UpdateNamespace)
				namespaces.DELETE("/:id", requireGlobalAdmin(), namespaceHandler.DeleteNamespace)
			}

			groups := authenticated.Group("/groups")
			{
				groups.GET("", groupViewerRBAC(), namespaceHandler.ListGroups)
				groups.POST("", groupEditorRBAC(), namespaceHandler.CreateGroup)
				groups.GET("/:id", groupByIDViewerRBAC(), namespaceHandler.GetGroup)
				groups.PUT("/:id", groupByIDEditorRBAC(), namespaceHandler.UpdateGroup)
				groups.DELETE("/:id", groupByIDEditorRBAC(), namespaceHandler.DeleteGroup)
			}

			configs := authenticated.Group("/configs")
			{
				configs.GET("", configViewerRBAC(), configHandler.ListConfigItems)
				configs.POST("", configEditorRBACByBody(), configHandler.CreateConfigItem)
				configs.GET("/:id", configByIDViewerRBAC(), configHandler.GetConfigItem)
				configs.PUT("/:id", configByIDEditorRBAC(), configHandler.UpdateConfigItem)
				configs.DELETE("/:id", configByIDEditorRBAC(), configHandler.DeleteConfigItem)
				configs.POST("/validate", configHandler.ValidateConfig)
				configs.GET("/merged", configViewerRBAC(), configHandler.GetMergedConfig)
				configs.POST("/:id/rollback", configByIDEditorRBAC(), configHandler.RollbackVersion)
				configs.GET("/:id/versions", configByIDViewerRBAC(), configHandler.GetVersionHistory)
				configs.GET("/:id/compare", configByIDViewerRBAC(), configHandler.CompareVersions)
				configs.POST("/batch-delete", configBatchDeleteRBAC(), configHandler.BatchDeleteConfigItems)
				configs.POST("/batch-copy", configBatchCopyRBAC(), configHandler.BatchCopyConfigItems)
			}

			gray := authenticated.Group("/gray")
			{
				gray.GET("", configGrayViewerRBAC(), grayHandler.ListGrayReleases)
				gray.POST("", configGrayEditorRBACByBody(), grayHandler.CreateGrayRelease)
				gray.GET("/:id", grayByIDViewerRBAC(), grayHandler.GetGrayRelease)
				gray.POST("/:id/start", grayByIDEditorRBAC(), grayHandler.StartGrayRelease)
				gray.POST("/:id/full-push", grayByIDEditorRBAC(), grayHandler.FullPush)
				gray.POST("/:id/rollback", grayByIDEditorRBAC(), grayHandler.RollbackGrayRelease)
			}

			push := authenticated.Group("/push")
			{
				push.GET("/long-poll", pushHandler.LongPoll)
				push.GET("/connections", pushHandler.GetConnections)
				push.GET("/stats", requireGlobalAdmin(), pushHandler.GetConnectionStats)
			}

			ws := authenticated.Group("/ws")
			{
				ws.GET("", wsHandler.HandleWebSocket)
			}

			metrics := authenticated.Group("/metrics")
			{
				metrics.GET("", requireGlobalAdmin(), metricHandler.GetMetrics)
				metrics.GET("/latest", requireGlobalAdmin(), metricHandler.GetLatestMetrics)
			}
		}
	}

	log.Printf("Server starting on port %d", cfg.ServerPort)
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func requireGlobalAdmin() gin.HandlerFunc {
	roleService := services.NewRoleService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		if !roleService.IsGlobalAdmin(userID) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			return
		}
		c.Next()
	}
}

func requireGlobalAdminOrSelf() gin.HandlerFunc {
	roleService := services.NewRoleService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		if roleService.IsGlobalAdmin(userID) {
			c.Next()
			return
		}
		queryUserID := c.Query("user_id")
		if queryUserID == "" {
			c.Next()
			return
		}
		qid, err := strconv.ParseUint(queryUserID, 10, 32)
		if err != nil {
			c.Next()
			return
		}
		if uint(qid) == userID {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin role required"})
	}
}

func namespaceViewerRBAC() gin.HandlerFunc {
	roleService := services.NewRoleService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.Next()
			return
		}
		nsID := uint(id)
		if !roleService.HasPermission(userID, nsID, models.RoleViewer) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

func groupViewerRBAC() gin.HandlerFunc {
	roleService := services.NewRoleService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, err := strconv.ParseUint(c.Query("namespace_id"), 10, 32)
		if err != nil {
			c.Next()
			return
		}
		nsID := uint(id)
		if !roleService.HasPermission(userID, nsID, models.RoleViewer) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

func groupByIDViewerRBAC() gin.HandlerFunc {
	roleService := services.NewRoleService()
	groupSvc := services.NewGroupService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.Next()
			return
		}
		group, err := groupSvc.GetGroup(uint(id))
		if err != nil {
			c.Next()
			return
		}
		if !roleService.HasPermission(userID, group.NamespaceID, models.RoleViewer) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

func groupEditorRBAC() gin.HandlerFunc {
	roleService := services.NewRoleService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		if roleService.IsGlobalAdmin(userID) {
			c.Next()
			return
		}
		body, ok := middleware.ParseJSONBody(c)
		if ok {
			if nsID, found := body["namespace_id"].(float64); found {
				if !roleService.HasPermission(userID, uint(nsID), models.RoleEditor) {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
					return
				}
			}
		}
		c.Next()
	}
}

func groupByIDEditorRBAC() gin.HandlerFunc {
	roleService := services.NewRoleService()
	groupSvc := services.NewGroupService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.Next()
			return
		}
		group, err := groupSvc.GetGroup(uint(id))
		if err != nil {
			c.Next()
			return
		}
		if !roleService.HasPermission(userID, group.NamespaceID, models.RoleEditor) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

func configViewerRBAC() gin.HandlerFunc {
	roleService := services.NewRoleService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, err := strconv.ParseUint(c.Query("namespace_id"), 10, 32)
		if err != nil {
			c.Next()
			return
		}
		nsID := uint(id)
		if !roleService.HasPermission(userID, nsID, models.RoleViewer) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

func configByIDViewerRBAC() gin.HandlerFunc {
	roleService := services.NewRoleService()
	configSvc := services.NewConfigService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.Next()
			return
		}
		item, err := configSvc.GetConfigItem(uint(id))
		if err != nil {
			c.Next()
			return
		}
		if !roleService.HasPermission(userID, item.NamespaceID, models.RoleViewer) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

func configEditorRBACByBody() gin.HandlerFunc {
	roleService := services.NewRoleService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		if roleService.IsGlobalAdmin(userID) {
			c.Next()
			return
		}
		body, ok := middleware.ParseJSONBody(c)
		if ok {
			if nsID, found := body["namespace_id"].(float64); found {
				if !roleService.HasPermission(userID, uint(nsID), models.RoleEditor) {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
					return
				}
			}
		}
		c.Next()
	}
}

func configByIDEditorRBAC() gin.HandlerFunc {
	roleService := services.NewRoleService()
	configSvc := services.NewConfigService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.Next()
			return
		}
		item, err := configSvc.GetConfigItem(uint(id))
		if err != nil {
			c.Next()
			return
		}
		if !roleService.HasPermission(userID, item.NamespaceID, models.RoleEditor) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

func configBatchDeleteRBAC() gin.HandlerFunc {
	roleService := services.NewRoleService()
	configSvc := services.NewConfigService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		if roleService.IsGlobalAdmin(userID) {
			c.Next()
			return
		}
		body, ok := middleware.ParseJSONBody(c)
		if ok {
			idsRaw, found := body["ids"].([]interface{})
			if !found {
				c.Next()
				return
			}
			for _, rawID := range idsRaw {
				idFloat, good := rawID.(float64)
				if !good {
					continue
				}
				item, err := configSvc.GetConfigItem(uint(idFloat))
				if err != nil {
					continue
				}
				if !roleService.HasPermission(userID, item.NamespaceID, models.RoleEditor) {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions on namespace " + strconv.Itoa(int(item.NamespaceID))})
					return
				}
			}
		}
		c.Next()
	}
}

func configBatchCopyRBAC() gin.HandlerFunc {
	roleService := services.NewRoleService()
	configSvc := services.NewConfigService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		if roleService.IsGlobalAdmin(userID) {
			c.Next()
			return
		}
		body, ok := middleware.ParseJSONBody(c)
		if ok {
			idsRaw, found := body["source_ids"].([]interface{})
			if !found {
				c.Next()
				return
			}
			for _, rawID := range idsRaw {
				idFloat, good := rawID.(float64)
				if !good {
					continue
				}
				item, err := configSvc.GetConfigItem(uint(idFloat))
				if err != nil {
					continue
				}
				if !roleService.HasPermission(userID, item.NamespaceID, models.RoleEditor) {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
					return
				}
			}
		}
		c.Next()
	}
}

func configGrayViewerRBAC() gin.HandlerFunc {
	roleService := services.NewRoleService()
	configSvc := services.NewConfigService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		idStr := c.Query("config_item_id")
		if idStr == "" {
			c.Next()
			return
		}
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.Next()
			return
		}
		item, err := configSvc.GetConfigItem(uint(id))
		if err != nil {
			c.Next()
			return
		}
		if !roleService.HasPermission(userID, item.NamespaceID, models.RoleViewer) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

func configGrayEditorRBACByBody() gin.HandlerFunc {
	roleService := services.NewRoleService()
	configSvc := services.NewConfigService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		if roleService.IsGlobalAdmin(userID) {
			c.Next()
			return
		}
		body, ok := middleware.ParseJSONBody(c)
		if ok {
			if cid, found := body["config_item_id"].(float64); found {
				item, err := configSvc.GetConfigItem(uint(cid))
				if err != nil {
					c.Next()
					return
				}
				if !roleService.HasPermission(userID, item.NamespaceID, models.RoleEditor) {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
					return
				}
			}
		}
		c.Next()
	}
}

func grayByIDViewerRBAC() gin.HandlerFunc {
	roleService := services.NewRoleService()
	graySvc := services.NewGrayReleaseService()
	configSvc := services.NewConfigService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.Next()
			return
		}
		gray, err := graySvc.GetGrayRelease(uint(id))
		if err != nil {
			c.Next()
			return
		}
		item, err := configSvc.GetConfigItem(gray.ConfigItemID)
		if err != nil {
			c.Next()
			return
		}
		if !roleService.HasPermission(userID, item.NamespaceID, models.RoleViewer) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

func grayByIDEditorRBAC() gin.HandlerFunc {
	roleService := services.NewRoleService()
	graySvc := services.NewGrayReleaseService()
	configSvc := services.NewConfigService()
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.Next()
			return
		}
		gray, err := graySvc.GetGrayRelease(uint(id))
		if err != nil {
			c.Next()
			return
		}
		item, err := configSvc.GetConfigItem(gray.ConfigItemID)
		if err != nil {
			c.Next()
			return
		}
		if !roleService.HasPermission(userID, item.NamespaceID, models.RoleEditor) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

func isWriteMethod(method string) bool {
	method = strings.ToUpper(method)
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete
}
