package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                     string
	DatabaseURL              string
	RedisURL                 string
	JWTSecret                string
	JWTExpiry                time.Duration
	FXProviderURL            string
	FXCacheDuration          time.Duration
	FXSpreadPct              float64
	QuoteExpiry              time.Duration
	Environment              string
	RateLimitConversionRPS   float64
	RateLimitConversionBurst int
	RateLimitPayoutRPS       float64
	RateLimitPayoutBurst     int
}

func Load() (*Config, error) {
	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "")
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	jwtExpiryHours, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	fxCacheMinutes, _ := strconv.Atoi(getEnv("FX_CACHE_MINUTES", "5"))
	spreadPct, _ := strconv.ParseFloat(getEnv("FX_SPREAD_PCT", "0.75"), 64)
	quoteExpirySecs, _ := strconv.Atoi(getEnv("QUOTE_EXPIRY_SECONDS", "30"))
	conversionRPS, _ := strconv.ParseFloat(getEnv("RATE_LIMIT_CONVERSION_RPS", "5"), 64)
	conversionBurst, _ := strconv.Atoi(getEnv("RATE_LIMIT_CONVERSION_BURST", "10"))
	payoutRPS, _ := strconv.ParseFloat(getEnv("RATE_LIMIT_PAYOUT_RPS", "3"), 64)
	payoutBurst, _ := strconv.Atoi(getEnv("RATE_LIMIT_PAYOUT_BURST", "5"))

	return &Config{
		Port:                     port,
		DatabaseURL:              dbURL,
		RedisURL:                 getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:                jwtSecret,
		JWTExpiry:                time.Duration(jwtExpiryHours) * time.Hour,
		FXProviderURL:            getEnv("FX_PROVIDER_URL", "https://api.frankfurter.app"),
		FXCacheDuration:          time.Duration(fxCacheMinutes) * time.Minute,
		FXSpreadPct:              spreadPct,
		QuoteExpiry:              time.Duration(quoteExpirySecs) * time.Second,
		Environment:              getEnv("ENVIRONMENT", "development"),
		RateLimitConversionRPS:   conversionRPS,
		RateLimitConversionBurst: conversionBurst,
		RateLimitPayoutRPS:       payoutRPS,
		RateLimitPayoutBurst:     payoutBurst,
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
