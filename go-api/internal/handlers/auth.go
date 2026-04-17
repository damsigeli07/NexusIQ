// auth.go
package handlers

import (
    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/damsigeli07/nexusiq/internal/config"
)

func Register(db *pgxpool.Pool) gin.HandlerFunc {
    return func(c *gin.Context) { c.JSON(200, gin.H{"message": "register - coming Day 2"}) }
}

func Login(db *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) { c.JSON(200, gin.H{"message": "login - coming Day 2"}) }
}
