// analytics.go
package handlers
import ("github.com/gin-gonic/gin"; "github.com/jackc/pgx/v5/pgxpool")
func GetAnalytics(db *pgxpool.Pool) gin.HandlerFunc {
    return func(c *gin.Context) { c.JSON(200, gin.H{"message": "analytics - coming Day 4"}) }
}
