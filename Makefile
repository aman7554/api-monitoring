.PHONY: build test run-api run-scheduler run-worker docker-up docker-down help

APP_NAME = pulsewatch

build:
	@echo "Building PulseWatch binaries..."
	go build -o bin/api ./cmd/api
	go build -o bin/scheduler ./cmd/scheduler
	go build -o bin/worker ./cmd/worker
	@echo "Build successful! Binaries generated in ./bin"

run-api:
	go run ./cmd/api/main.go

run-scheduler:
	go run ./cmd/scheduler/main.go

run-worker:
	go run ./cmd/worker/main.go

test:
	go test -v ./...

docker-up:
	docker-compose -f deployments/docker/docker-compose.yml up --build -d

docker-down:
	docker-compose -f deployments/docker/docker-compose.yml down -v

helm-lint:
	helm lint deployments/helm/pulsewatch

help:
	@echo "PulseWatch Makefile Commands:"
	@echo "  make build          - Build API, Scheduler, Worker binaries"
	@echo "  make run-api        - Run API server locally"
	@echo "  make run-scheduler  - Run Scheduler service locally"
	@echo "  make run-worker     - Run Worker service locally"
	@echo "  make test           - Run all unit and integration tests"
	@echo "  make docker-up      - Spin up local environment with Docker Compose"
	@echo "  make docker-down    - Stop and clean up Docker Compose environment"
