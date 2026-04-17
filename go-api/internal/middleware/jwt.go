package middleware

import (
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "github.com/damsigeli07/nexusiq/internal/config"
)

type Claims struct {
    UserID   string `json:"user_id"`
    TenantID string `json:"tenant_id"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

func JWTAuth(cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        header := c.GetHeader("Authorization")
        if !strings.HasPrefix(header, "Bearer ") {
            c.AbortWithStatusJSON(401, gin.H{"error": "missing token"})
            return
        }

        tokenStr := strings.TrimPrefix(header, "Bearer ")
        claims := &Claims{}

        token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
            return []byte(cfg.JWTSecret), nil
        })

        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
            return
        }

        // Inject into context — handlers pull from here
        c.Set("user_id", claims.UserID)
        c.Set("tenant_id", claims.TenantID)
        c.Set("role", claims.Role)
        c.Next()
    }
}
