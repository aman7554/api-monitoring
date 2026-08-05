package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"pulsewatch/internal/config"
	"pulsewatch/internal/domain"
	"pulsewatch/internal/queue"
	"pulsewatch/internal/repository/postgres"
	"pulsewatch/internal/service"
	"pulsewatch/internal/service/checker"
	"pulsewatch/internal/telemetry"
)

type WorkerPool struct {
	concurrency int
	redisQ      *queue.RedisQueue
	monRepo     *postgres.MonitorRepository
	checkRepo   *postgres.CheckRepository
	incSvc      *service.IncidentService
	httpChecker *checker.HTTPChecker
	sslChecker  *checker.SSLChecker
	dnsChecker  *checker.DNSChecker
	wg          sync.WaitGroup
	stopChan    chan struct{}
}

func NewWorkerPool(
	concurrency int,
	redisQ *queue.RedisQueue,
	monRepo *postgres.MonitorRepository,
	checkRepo *postgres.CheckRepository,
	incSvc *service.IncidentService,
) *WorkerPool {
	return &WorkerPool{
		concurrency: concurrency,
		redisQ:      redisQ,
		monRepo:     monRepo,
		checkRepo:   checkRepo,
		incSvc:      incSvc,
		httpChecker: checker.NewHTTPChecker(),
		sslChecker:  checker.NewSSLChecker(),
		dnsChecker:  checker.NewDNSChecker(),
		stopChan:    make(chan struct{}),
	}
}

func (p *WorkerPool) Start() {
	log.Printf("[PulseWatch Worker] Starting pool with %d concurrent workers...\n", p.concurrency)
	for i := 0; i < p.concurrency; i++ {
		p.wg.Add(1)
		go p.workerRoutine(i + 1)
	}
}

func (p *WorkerPool) Stop() {
	close(p.stopChan)
	p.wg.Wait()
	log.Println("[PulseWatch Worker] All workers stopped cleanly.")
}

func (p *WorkerPool) workerRoutine(id int) {
	defer p.wg.Done()
	log.Printf("[PulseWatch Worker #%d] Worker ready for jobs.\n", id)

	for {
		select {
		case <-p.stopChan:
			return
		default:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			job, err := p.redisQ.DequeueCheckJob(ctx, 2*time.Second)
			cancel()

			if err != nil || job == nil {
				continue
			}

			p.processJob(job)
		}
	}
}

func (p *WorkerPool) processJob(job *queue.CheckJob) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	m, err := p.monRepo.GetByID(ctx, job.MonitorID)
	if err != nil {
		log.Printf("[PulseWatch Worker] Monitor ID %s not found: %v", job.MonitorID, err)
		return
	}

	if !m.IsActive {
		return
	}

	log.Printf("[PulseWatch Worker] Executing check for monitor '%s' (%s, type=%s)...", m.Name, m.URL, m.Type)

	var detail *checker.CheckResultDetail
	switch m.Type {
	case domain.MonitorTypeSSL:
		detail = p.sslChecker.Execute(ctx, m)
	case domain.MonitorTypeDNS:
		detail = p.dnsChecker.Execute(ctx, m)
	default:
		detail = p.httpChecker.Execute(ctx, m)
	}

	res := &domain.CheckResult{
		MonitorID:        m.ID,
		Status:           detail.Status,
		StatusCode:       detail.StatusCode,
		LatencyMS:        detail.LatencyMS,
		DNSTimeMS:        detail.DNSTimeMS,
		SSLDaysRemaining: detail.SSLDaysRemaining,
		ErrorMessage:     detail.ErrorMessage,
		CheckedAt:        time.Now(),
	}

	// Persist Check Result
	if err := p.checkRepo.Create(ctx, res); err != nil {
		log.Printf("[PulseWatch Worker] Failed to save check result for monitor %s: %v", m.ID, err)
	}

	// Update Incident State & Monitor Status
	if err := p.incSvc.ProcessCheckResult(ctx, m, res); err != nil {
		log.Printf("[PulseWatch Worker] Failed to update incident state for monitor %s: %v", m.ID, err)
	}

	// Prometheus telemetry
	telemetry.MonitorChecksTotal.WithLabelValues(string(m.Type), string(res.Status)).Inc()
	telemetry.MonitorLatencyMS.WithLabelValues(string(m.Type), string(res.Status)).Observe(float64(res.LatencyMS))
}

func main() {
	cfg := config.LoadConfig()
	fmt.Println("[PulseWatch Worker] Starting Monitor Worker Service...")

	db, err := postgres.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[PulseWatch Worker] DB Connection Error: %v", err)
	}
	defer db.Close()

	redisQ, err := queue.NewRedisQueue(cfg.RedisURL)
	if err != nil {
		log.Fatalf("[PulseWatch Worker] Redis Connection Error: %v", err)
	}
	defer redisQ.Close()

	monRepo := postgres.NewMonitorRepository(db)
	checkRepo := postgres.NewCheckRepository(db)
	incRepo := postgres.NewIncidentRepository(db)
	notifRepo := postgres.NewNotificationRepository(db)

	notifSvc := service.NewNotificationService(notifRepo)
	incSvc := service.NewIncidentService(incRepo, monRepo, notifSvc)

	pool := NewWorkerPool(cfg.WorkerConcurrency, redisQ, monRepo, checkRepo, incSvc)
	pool.Start()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("[PulseWatch Worker] Serving Prometheus metrics on :8081/metrics")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Printf("[PulseWatch Worker] Metrics HTTP server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[PulseWatch Worker] Signal received, shutting down worker pool...")
	pool.Stop()
}
