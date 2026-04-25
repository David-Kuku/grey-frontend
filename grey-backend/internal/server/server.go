package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"

	"github.com/David-Kuku/grey-frontend/grey-backend/internal/auth"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/config"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/fx"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/handlers"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/ledger"
	appMiddleware "github.com/David-Kuku/grey-frontend/grey-backend/internal/middleware"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/payout"
	"github.com/David-Kuku/grey-frontend/grey-backend/internal/repository"
)

type Server struct {
	Router *chi.Mux
	DB     *sqlx.DB
	Redis  *redis.Client
	Config *config.Config
	Logger *slog.Logger
}

func New(db *sqlx.DB, redisClient *redis.Client, cfg *config.Config, logger *slog.Logger) *Server {
	s := &Server{
		Router: chi.NewRouter(),
		DB:     db,
		Redis:  redisClient,
		Config: cfg,
		Logger: logger,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.Router.Use(chimiddleware.Recoverer)
	s.Router.Use(chimiddleware.RealIP)
	s.Router.Use(appMiddleware.RequestID)
	s.Router.Use(appMiddleware.Logger(s.Logger))
	s.Router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", "X-Idempotency-Key", "Idempotency-Key"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

func (s *Server) setupRoutes() {
	repo := repository.New(s.DB)
	authService := auth.NewService(s.Config.JWTSecret, s.Config.JWTExpiry)
	ledgerService := ledger.NewService(repo)
	fxService := fx.NewService(repo, s.Config, s.Redis)
	payoutService, err := payout.NewService(repo, ledgerService, s.Logger, s.Redis)
	if err != nil {
		s.Logger.Error("failed to create payout service", "error", err)
		panic(err)
	}
	workerRedisOpts := s.Redis.Options()
	workerRedisClient := redis.NewClient(workerRedisOpts)
	go func() {
		if err := payoutService.StartWorker(context.Background(), workerRedisClient); err != nil {
			s.Logger.Error("payout worker stopped", "error", err)
		}
	}()
	authHandler := handlers.NewAuthHandler(repo, authService)
	depositHandler := handlers.NewDepositHandler(repo, ledgerService)
	conversionHandler := handlers.NewConversionHandler(repo, fxService, ledgerService)
	payoutHandler := handlers.NewPayoutHandler(repo, ledgerService, payoutService, s.Logger)
	walletHandler := handlers.NewWalletHandler(repo)
	conversionLimiter := appMiddleware.NewPerUserRateLimiter(appMiddleware.RateLimiterConfig{
		Rate:  rate.Limit(s.Config.RateLimitConversionRPS),
		Burst: s.Config.RateLimitConversionBurst,
	})
	payoutLimiter := appMiddleware.NewPerUserRateLimiter(appMiddleware.RateLimiterConfig{
		Rate:  rate.Limit(s.Config.RateLimitPayoutRPS),
		Burst: s.Config.RateLimitPayoutBurst,
	})
	s.Router.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"ok"}`))
		})
		r.Post("/auth/signup", authHandler.Signup)
		r.Post("/auth/login", authHandler.Login)
		r.Group(func(r chi.Router) {
			r.Use(appMiddleware.Auth(authService))
			r.Get("/wallet/balances", walletHandler.GetBalances)
			r.Post("/deposits", depositHandler.CreateDeposit)
			r.Group(func(r chi.Router) {
				r.Use(conversionLimiter.Middleware)
				r.Post("/conversions/quote", conversionHandler.GetQuote)
				r.Post("/conversions/execute", conversionHandler.ExecuteConversion)
			})
			r.Group(func(r chi.Router) {
				r.Use(payoutLimiter.Middleware)
				r.Post("/payouts", payoutHandler.CreatePayout)
			})
			r.Get("/transactions", walletHandler.GetTransactions)
			r.Get("/transactions/{id}", walletHandler.GetTransaction)
		})
	})
}

func (s *Server) Start() error {
	addr := ":" + s.Config.Port
	s.Logger.Info("server starting", "addr", addr, "env", s.Config.Environment)
	return http.ListenAndServe(addr, s.Router)
}