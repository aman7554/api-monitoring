package handler

import (
	"database/sql"
	"net/http"

	"pulsewatch/internal/middleware"
	"pulsewatch/internal/queue"
	"pulsewatch/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type RouterDependencies struct {
	DB               *sql.DB
	RedisQ           *queue.RedisQueue
	AuthSvc          *service.AuthService
	OrgSvc           *service.OrgService
	ProjSvc          *service.ProjectService
	MonSvc           *service.MonitorService
	DashSvc          *service.DashboardService
	StatusSvc        *service.StatusPageService
	AuditSvc         *service.AuditService
	AuthHandler      *AuthHandler
	OrgHandler       *OrgHandler
	ProjectHandler   *ProjectHandler
	MonitorHandler   *MonitorHandler
	DashboardHandler *DashboardHandler
	StatusHandler    *StatusHandler
	ApiKeyHandler    *ApiKeyHandler
	AuditHandler     *AuditHandler
}

func SetupRouter(deps *RouterDependencies) *gin.Engine {
	r := gin.New()

	// Middlewares
	r.Use(gin.Recovery())
	r.Use(middleware.StructuredLogger())
	r.Use(middleware.PrometheusMiddleware())

	// Health Checks
	r.GET("/healthz", func(c *gin.Context) {
		if err := deps.DB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "db": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "pulsewatch-api"})
	})

	r.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// Prometheus Metrics Endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Serve Live Web Dashboard UI
	r.Static("/web", "web")
	r.StaticFile("/", "web/index.html")
	r.StaticFile("/dashboard", "web/index.html")

	// Serve OpenAPI Spec & Swagger UI
	r.StaticFile("/docs/swagger.yaml", "docs/swagger.yaml")
	r.GET("/swagger/*any", func(c *gin.Context) {
		if c.Param("any") == "/swagger.yaml" || c.Param("any") == "/doc.json" {
			c.File("docs/swagger.yaml")
			return
		}
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, swaggerUIHTML)
	})

	// Public Routes
	v1 := r.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", deps.AuthHandler.Register)
			authGroup.POST("/login", deps.AuthHandler.Login)
		}

		publicGroup := v1.Group("/public")
		{
			publicGroup.GET("/status/:slug", deps.StatusHandler.GetPublicStatus)
		}
	}

	// Protected Routes (JWT / API Key)
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(deps.AuthSvc))
	if deps.RedisQ != nil {
		protected.Use(middleware.RateLimitMiddleware(deps.RedisQ, 100, 60))
	}
	{
		protected.GET("/auth/me", deps.AuthHandler.Me)

		// Organizations
		protected.POST("/organizations", deps.OrgHandler.CreateOrg)
		protected.GET("/organizations", deps.OrgHandler.ListUserOrgs)
		protected.GET("/organizations/:org_id/audit-logs", middleware.RequireOrgRole(deps.OrgSvc, "member"), deps.AuditHandler.ListOrgLogs)

		// Projects
		protected.POST("/projects", deps.ProjectHandler.CreateProject)
		protected.GET("/projects", deps.ProjectHandler.ListOrgProjects)
		protected.GET("/projects/:id", deps.ProjectHandler.GetProject)
		protected.GET("/projects/:id/dashboard", deps.DashboardHandler.GetProjectDashboard)

		// Monitors
		protected.POST("/monitors", deps.MonitorHandler.CreateMonitor)
		protected.GET("/monitors", deps.MonitorHandler.ListProjectMonitors)
		protected.GET("/monitors/:id", deps.MonitorHandler.GetMonitor)
		protected.DELETE("/monitors/:id", deps.MonitorHandler.DeleteMonitor)
		protected.GET("/monitors/:id/checks", deps.MonitorHandler.ListMonitorChecks)

		// API Keys
		protected.POST("/api-keys", deps.ApiKeyHandler.CreateApiKey)
		protected.GET("/api-keys", deps.ApiKeyHandler.ListProjectKeys)
	}

	return r
}

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>PulseWatch API Documentation</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin:0; background: #fafafa; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" charset="UTF-8"> </script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js" charset="UTF-8"> </script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "/docs/swagger.yaml",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      });
    };
  </script>
</body>
</html>`

