module github.com/David-Kuku/grey-frontend/grey-backend

go 1.22

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/go-chi/cors v1.2.1
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/google/uuid v1.6.0
	github.com/jmoiron/sqlx v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.18.0
	go.codycody31.dev/gobullmq v0.0.0-20260316234017-9f1b90a49da7
	golang.org/x/crypto v0.25.0
	golang.org/x/time v0.5.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
)

replace go.codycody31.dev/gobullmq => github.com/Codycody31/gobullmq v0.0.0-20260316234017-9f1b90a49da7
