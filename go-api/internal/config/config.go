package config

import "os"

type Config struct {
    Port          string
    DatabaseURL   string
    RedisAddr     string
    RedisPassword string
    JWTSecret     string
    MLServiceURL  string
}

func Load() *Config {
    return &Config{
        Port: getEnv("PORT", "8080"),
        DatabaseURL: "postgres://" + os.Getenv("POSTGRES_USER") +
            ":" + os.Getenv("POSTGRES_PASSWORD") +
            "@postgres:5432/" + os.Getenv("POSTGRES_DB") + "?sslmode=disable",
        RedisAddr:     "redis:6379",
        RedisPassword: os.Getenv("REDIS_PASSWORD"),
        JWTSecret:     os.Getenv("JWT_SECRET"),
        MLServiceURL:  "http://python-ml:8000",
    }
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
