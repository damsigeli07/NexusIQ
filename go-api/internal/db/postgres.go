package db

import (
    "context"
    "log"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/redis/go-redis/v9"
    "github.com/damsigeli07/NexusIQ/internal/config"
)

func Connect(cfg *config.Config) *pgxpool.Pool {
    pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("Unable to connect to database: %v", err)
    }
    log.Println("Connected to PostgreSQL")
    return pool
}

func ConnectRedis(cfg *config.Config) *redis.Client {
    client := redis.NewClient(&redis.Options{
        Addr:     cfg.RedisAddr,
        Password: cfg.RedisPassword,
    })
    log.Println("Connected to Redis")
    return client
}
