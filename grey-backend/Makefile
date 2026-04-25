.PHONY: help dev test migrate build docker-up docker-down clean

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
dev:
	go run ./cmd/api

build:
	CGO_ENABLED=0 go build -o bin/kite-api ./cmd/api
docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f
migrate:
	psql "$(DATABASE_URL)" -f migrations/001_initial_schema.sql

migrate-test:
	psql "$(TEST_DATABASE_URL)" -f migrations/001_initial_schema.sql
test:
	go test -v -count=1 -timeout 60s ./tests/...

test-race:
	go test -v -race -count=1 -timeout 120s ./tests/...

test-cover:
	go test -v -coverprofile=coverage.out ./tests/... ./internal/...
	go tool cover -html=coverage.out -o coverage.html
lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ coverage.out coverage.html

seed:
	@echo "Creating test user..."
	@curl -s -X POST http://localhost:8080/api/v1/auth/signup \
		-H "Content-Type: application/json" \
		-d '{"email":"demo@kite.test","password":"password123"}' | jq .
	@echo "\nDepositing funds..."
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
		-H "Content-Type: application/json" \
		-d '{"email":"demo@kite.test","password":"password123"}' | jq -r .token); \
	curl -s -X POST http://localhost:8080/api/v1/deposits \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"currency":"USD","amount":1000000,"idempotency_key":"seed-usd"}' | jq .; \
	curl -s -X POST http://localhost:8080/api/v1/deposits \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"currency":"NGN","amount":50000000,"idempotency_key":"seed-ngn"}' | jq .; \
	curl -s -X POST http://localhost:8080/api/v1/deposits \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"currency":"GBP","amount":500000,"idempotency_key":"seed-gbp"}' | jq .; \
	curl -s -X POST http://localhost:8080/api/v1/deposits \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"currency":"EUR","amount":750000,"idempotency_key":"seed-eur"}' | jq .; \
	curl -s -X POST http://localhost:8080/api/v1/deposits \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer $$TOKEN" \
		-d '{"currency":"KES","amount":10000000,"idempotency_key":"seed-kes"}' | jq .
	@echo "\nDone! Login: demo@kite.test / password123"
