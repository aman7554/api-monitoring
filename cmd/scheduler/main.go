package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pulsewatch/internal/config"
	"pulsewatch/internal/queue"
	"pulsewatch/internal/repository/postgres"
)

func main() {
	cfg := config.LoadConfig()
	fmt.Println("[PulseWatch Scheduler] Starting Monitor Scheduler Service...")

	db, err := postgres.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[PulseWatch Scheduler] DB Connection Error: %v", err)
	}
	defer db.Close()

	redisQ, err := queue.NewRedisQueue(cfg.RedisURL)
	if err != nil {
		log.Fatalf("[PulseWatch Scheduler] Redis Connection Error: %v", err)
	}
	defer redisQ.Close()

	monRepo := postgres.NewMonitorRepository(db)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("[PulseWatch Scheduler] Scheduler running. Polling for due monitors every 5 seconds...")

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			monitors, err := monRepo.ListDueMonitors(ctx, 100)
			cancel()

			if err != nil {
				log.Printf("[PulseWatch Scheduler] Error querying due monitors: %v", err)
				continue
			}

			if len(monitors) > 0 {
				log.Printf("[PulseWatch Scheduler] Found %d due monitor(s). Enqueuing...", len(monitors))
				for _, m := range monitors {
					eCtx, eCancel := context.WithTimeout(context.Background(), 3*time.Second)
					if err := redisQ.EnqueueCheckJob(eCtx, m.ID); err != nil {
						log.Printf("[PulseWatch Scheduler] Failed to enqueue monitor %s: %v", m.ID, err)
					}
					eCancel()
				}
			}
		case <-stopChan:
			log.Println("[PulseWatch Scheduler] Shutting down scheduler gracefully...")
			return
		}
	}
}
