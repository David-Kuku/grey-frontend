package fx

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/config"
	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/models"
	"github.com/David-Kuku/kuku-kite-app/grey-backend/internal/repository"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	repo       *repository.Repository
	cfg        *config.Config
	redis      *redis.Client
	httpClient *http.Client
}

func NewService(repo *repository.Repository, cfg *config.Config, redisClient *redis.Client) *Service {
	return &Service{
		repo:  repo,
		cfg:   cfg,
		redis: redisClient,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type frankfurterResponse struct {
	Date  string  `json:"date"`
	Base  string  `json:"base"`
	Quote string  `json:"quote"`
	Rate  float64 `json:"rate"`
}

func (s *Service) GetRate(ctx context.Context, base, target models.Currency) (float64, error) {
	key := rateCacheKey(base, target)
	cached, err := s.redis.Get(ctx, key).Result()
	if err == nil {
		rate, parseErr := strconv.ParseFloat(cached, 64)
		if parseErr != nil {
			return 0, fmt.Errorf("parse cached rate: %w", parseErr)
		}
		return rate, nil
	}
	if err != redis.Nil {
		return 0, fmt.Errorf("check redis rate cache: %w", err)
	}
	rate, err := s.fetchRate(ctx, base, target)
	if err != nil {
		return 0, err
	}
	if err := s.redis.Set(ctx, key, strconv.FormatFloat(rate, 'f', -1, 64), s.cfg.FXCacheDuration).Err(); err != nil {
		fmt.Printf("WARN: failed to cache rate in redis: %v\n", err)
	}

	return rate, nil
}

func rateCacheKey(base, target models.Currency) string {
	return fmt.Sprintf("fx_rate:%s:%s", base, target)
}
func (s *Service) fetchRate(ctx context.Context, base, target models.Currency) (float64, error) {
	url := fmt.Sprintf("%s?base=%s&quotes=%s", s.cfg.FXProviderURL, base, target)

	fmt.Println(url)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fetch rate from provider: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("provider returned status %d", resp.StatusCode)
	}

	var results []frankfurterResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return 0, fmt.Errorf("decode provider response: %w", err)
	}

	if len(results) == 0 {
		return 0, fmt.Errorf("rate not found for %s→%s", base, target)
	}

	return results[0].Rate, nil
}
func (s *Service) ApplySpread(marketRate float64) float64 {
	spreadMultiplier := 1 - (s.cfg.FXSpreadPct / 100)
	return marketRate * spreadMultiplier
}
func (s *Service) CreateQuote(ctx context.Context, userID uuid.UUID, req models.QuoteRequest) (*models.FXQuote, error) {
	if req.SourceCurrency == req.TargetCurrency {
		return nil, fmt.Errorf("source and target currencies must be different")
	}
	if req.SourceAmount <= 0 {
		return nil, fmt.Errorf("source amount must be positive")
	}

	marketRate, err := s.GetRate(ctx, req.SourceCurrency, req.TargetCurrency)
	if err != nil {
		return nil, fmt.Errorf("get market rate: %w", err)
	}

	quotedRate := s.ApplySpread(marketRate)
	targetAmount := int64(math.Floor(float64(req.SourceAmount) * quotedRate))
	marketTargetAmount := int64(math.Floor(float64(req.SourceAmount) * marketRate))
	feeAmount := marketTargetAmount - targetAmount

	feeCurrency := req.TargetCurrency

	quote := &models.FXQuote{
		ID:             uuid.New(),
		UserID:         userID,
		SourceCurrency: req.SourceCurrency,
		TargetCurrency: req.TargetCurrency,
		MarketRate:     fmt.Sprintf("%.10f", marketRate),
		QuotedRate:     fmt.Sprintf("%.10f", quotedRate),
		SpreadPct:      fmt.Sprintf("%.4f", s.cfg.FXSpreadPct),
		SourceAmount:   req.SourceAmount,
		TargetAmount:   targetAmount,
		FeeAmount:      feeAmount,
		FeeCurrency:    &feeCurrency,
		ExpiresAt:      time.Now().Add(s.cfg.QuoteExpiry),
	}

	if err := s.repo.CreateFXQuote(ctx, quote); err != nil {
		return nil, fmt.Errorf("save quote: %w", err)
	}

	return quote, nil
}
func (s *Service) GetQuote(ctx context.Context, quoteID uuid.UUID, userID uuid.UUID) (*models.FXQuote, error) {
	quote, err := s.repo.GetFXQuote(ctx, quoteID)
	if err != nil {
		return nil, fmt.Errorf("get quote: %w", err)
	}
	if quote.UserID != userID {
		return nil, fmt.Errorf("quote does not belong to this user")
	}
	return quote, nil
}
