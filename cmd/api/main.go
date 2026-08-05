package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pulsewatch/internal/config"
	"pulsewatch/internal/handler"
	"pulsewatch/internal/queue"
	"pulsewatch/internal/repository/postgres"
	"pulsewatch/internal/service"
	"pulsewatch/internal/telemetry"
)

func main() {
	cfg := config.LoadConfig()
	fmt.Printf("[PulseWatch API] Starting in %s mode on port %s...\n", cfg.Environment, cfg.Port)

	// Initialize OpenTelemetry
	tp, err := telemetry.InitTracer("pulsewatch-api")
	if err != nil {
		log.Printf("[PulseWatch API] Warning: OpenTelemetry init failed: %v", err)
	}
	defer telemetry.ShutdownTracer(context.Background(), tp)

	// Initialize DB
	db, err := postgres.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[PulseWatch API] DB Connection Error: %v", err)
	}
	defer db.Close()

	// Initialize Redis Queue
	redisQ, err := queue.NewRedisQueue(cfg.RedisURL)
	if err != nil {
		log.Printf("[PulseWatch API] Warning: Redis Queue init error: %v", err)
	} else {
		defer redisQ.Close()
	}

	// Repositories
	userRepo := postgres.NewUserRepository(db)
	orgRepo := postgres.NewOrgRepository(db)
	projRepo := postgres.NewProjectRepository(db)
	monRepo := postgres.NewMonitorRepository(db)
	checkRepo := postgres.NewCheckRepository(db)
	incRepo := postgres.NewIncidentRepository(db)
	keyRepo := postgres.NewApiKeyRepository(db)
	auditRepo := postgres.NewAuditRepository(db)

	// Services
	authSvc := service.NewAuthService(userRepo, keyRepo, cfg)
	orgSvc := service.NewOrgService(orgRepo)
	projSvc := service.NewProjectService(projRepo)
	monSvc := service.NewMonitorService(monRepo)
	dashSvc := service.NewDashboardService(monRepo, checkRepo, incRepo)
	statusSvc := service.NewStatusPageService(projRepo, monRepo, incRepo, checkRepo)
	auditSvc := service.NewAuditService(auditRepo)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	orgH := handler.NewOrgHandler(orgSvc, auditSvc)
	projH := handler.NewProjectHandler(projSvc, auditSvc)
	monH := handler.NewMonitorHandler(monSvc, checkRepo, auditSvc)
	dashH := handler.NewDashboardHandler(dashSvc)
	statusH := handler.NewStatusHandler(statusSvc)
	keyH := handler.NewApiKeyHandler(authSvc, keyRepo)
	auditH := handler.NewAuditHandler(auditSvc)

	deps := &handler.RouterDependencies{
		DB:               db,
		RedisQ:           redisQ,
		AuthSvc:          authSvc,
		OrgSvc:           orgSvc,
		ProjSvc:          projSvc,
		MonSvc:           monSvc,
		DashSvc:          dashSvc,
		StatusSvc:        statusSvc,
		AuditSvc:         auditSvc,
		AuthHandler:      authH,
		OrgHandler:       orgH,
		ProjectHandler:   projH,
		MonitorHandler:   monH,
		DashboardHandler: dashH,
		StatusHandler:    statusH,
		ApiKeyHandler:    keyH,
		AuditHandler:     auditH,
	}

	router := handler.SetupRouter(deps)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful Shutdown Channel
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[PulseWatch API] Listen error: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[PulseWatch API] Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("[PulseWatch API] Server forced to shutdown: %v", err)
	}

	log.Println("[PulseWatch API] Server exited cleanly.")
}
