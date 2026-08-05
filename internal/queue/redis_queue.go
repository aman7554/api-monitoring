package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const CheckQueueName = "pulsewatch:monitors:queue"

type CheckJob struct {
	MonitorID uuid.UUID `json:"monitor_id"`
	EnqueuedAt time.Time `json:"enqueued_at"`
	Attempt    int       `json:"attempt"`
}

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue(redisURL string) (*RedisQueue, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}

	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		// Log warning but allow client creation if starting up before redis
		fmt.Printf("[RedisQueue] Warning: ping failed: %v\n", err)
	}

	return &RedisQueue{client: client}, nil
}

func (q *RedisQueue) EnqueueCheckJob(ctx context.Context, monitorID uuid.UUID) error {
	job := CheckJob{
		MonitorID: monitorID,
		EnqueuedAt: time.Now(),
		Attempt:    1,
	}
	bytes, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return q.client.RPush(ctx, CheckQueueName, string(bytes)).Err()
}

func (q *RedisQueue) DequeueCheckJob(ctx context.Context, timeout time.Duration) (*CheckJob, error) {
	res, err := q.client.BLPop(ctx, timeout, CheckQueueName).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // timeout, queue empty
		}
		return nil, err
	}

	if len(res) < 2 {
		return nil, fmt.Errorf("invalid queue response format")
	}

	var job CheckJob
	if err := json.Unmarshal([]byte(res[1]), &job); err != nil {
		return nil, err
	}

	return &job, nil
}

func (q *RedisQueue) AllowRateLimit(ctx context.Context, key string, maxRequests int, window time.Duration) (bool, error) {
	redisKey := fmt.Sprintf("ratelimit:%s", key)
	count, err := q.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return true, nil // Fallback allow on Redis error
	}

	if count == 1 {
		q.client.Expire(ctx, redisKey, window)
	}

	return count <= int64(maxRequests), nil
}

func (q *RedisQueue) Close() error {
	return q.client.Close()
}
