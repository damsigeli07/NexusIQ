// documents.go
package handlers
import ("github.com/gin-gonic/gin"; "github.com/jackc/pgx/v5/pgxpool")
func ListDocuments(db *pgxpool.Pool) gin.HandlerFunc {
    return func(c *gin.Context) { c.JSON(200, gin.H{"documents": []string{}}) }
}
func UploadDocument(db *pgxpool.Pool) gin.HandlerFunc {
    return func(c *gin.Context) { c.JSON(200, gin.H{"message": "upload - coming Day 2"}) }
}
func DeleteDocument(db *pgxpool.Pool) gin.HandlerFunc {
    return func(c *gin.Context) { c.JSON(200, gin.H{"message": "deleted"}) }
}
