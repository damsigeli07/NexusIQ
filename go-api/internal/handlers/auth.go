package handlers

import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "golang.org/x/crypto/bcrypt"
    "github.com/damsigeli07/NexusIQ/internal/config"
    "github.com/damsigeli07/NexusIQ/internal/middleware"
)

type registerRequest struct {
    Email      string `json:"email"       binding:"required,email"`
    Password   string `json:"password"    binding:"required,min=6"`
    TenantSlug string `json:"tenant_slug" binding:"required"`
}

type loginRequest struct {
    Email    string `json:"email"    binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

func Register(db *pgxpool.Pool) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req registerRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        var tenantID string
        err := db.QueryRow(context.Background(),
            "SELECT id FROM tenants WHERE slug = $1", req.TenantSlug,
        ).Scan(&tenantID)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "tenant not found"})
            return
        }
        hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
            return
        }
        var userID string
        err = db.QueryRow(context.Background(),
            `INSERT INTO users (tenant_id, email, password_hash, role)
             VALUES ($1, $2, $3, 'member') RETURNING id`,
            tenantID, req.Email, string(hash),
        ).Scan(&userID)
        if err != nil {
            c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
            return
        }
        c.JSON(http.StatusCreated, gin.H{"message": "registered", "user_id": userID})
    }
}

func Login(db *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req loginRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        var userID, tenantID, role, hash string
        err := db.QueryRow(context.Background(),
            `SELECT id, tenant_id, role, password_hash
             FROM users WHERE email = $1`, req.Email,
        ).Scan(&userID, &tenantID, &role, &hash)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
            return
        }
        if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
            return
        }
        claims := &middleware.Claims{
            UserID:   userID,
            TenantID: tenantID,
            Role:     role,
            RegisteredClaims: jwt.RegisteredClaims{
                ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
                IssuedAt:  jwt.NewNumericDate(time.Now()),
            },
        }
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
        signed, err := token.SignedString([]byte(cfg.JWTSecret))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "could not sign token"})
            return
        }
        c.JSON(http.StatusOK, gin.H{
            "token":     signed,
            "user_id":   userID,
            "tenant_id": tenantID,
            "role":      role,
        })
    }
}
