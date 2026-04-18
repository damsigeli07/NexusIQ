// documents.go
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/damsigeli07/NexusIQ/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

const uploadsDir = "/app/uploads"

type documentResponse struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	SourceType string    `json:"source_type"`
	Status     string    `json:"status"`
	FilePath   string    `json:"file_path"`
	CreatedAt  time.Time `json:"created_at"`
}

type embedRequest struct {
	DocumentID string `json:"document_id"`
	TenantID   string `json:"tenant_id"`
	FilePath   string `json:"file_path"`
	SourceType string `json:"source_type"`
}

func inferSourceType(filename, override string) string {
	if override != "" {
		return override
	}
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".pdf":
		return "pdf"
	case ".docx":
		return "docx"
	default:
		return "txt"
	}
}

func ListDocuments(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		rows, err := db.Query(context.Background(),
			`SELECT id, title, source_type, status, file_path, created_at
             FROM documents WHERE tenant_id = $1 ORDER BY created_at DESC`,
			tenantID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not query documents"})
			return
		}
		defer rows.Close()

		docs := []documentResponse{}
		for rows.Next() {
			var doc documentResponse
			if err := rows.Scan(&doc.ID, &doc.Title, &doc.SourceType, &doc.Status, &doc.FilePath, &doc.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "could not read documents"})
				return
			}
			docs = append(docs, doc)
		}

		c.JSON(http.StatusOK, gin.H{"documents": docs})
	}
}

func UploadDocument(db *pgxpool.Pool, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		userID := c.GetString("user_id")

		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
			return
		}

		title := c.PostForm("title")
		if title == "" {
			title = file.Filename
		}
		sourceType := inferSourceType(file.Filename, c.PostForm("source_type"))

		if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create upload directory"})
			return
		}

		savedName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
		filePath := filepath.Join(uploadsDir, savedName)
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save file"})
			return
		}

		var documentID string
		err = db.QueryRow(context.Background(),
			`INSERT INTO documents (tenant_id, title, source_type, file_path, status, uploaded_by)
             VALUES ($1, $2, $3, $4, 'processing', $5) RETURNING id`,
			tenantID, title, sourceType, filePath, userID,
		).Scan(&documentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create document record"})
			return
		}

		// Ask Python ML service to ingest and embed the document.
		reqBody := embedRequest{
			DocumentID: documentID,
			TenantID:   tenantID,
			FilePath:   filePath,
			SourceType: sourceType,
		}
		bodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not serialize request"})
			return
		}

		resp, err := http.Post(cfg.MLServiceURL+"/embed", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "could not reach ML service"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusBadGateway, gin.H{"error": "ML service failed to process document"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "document uploaded", "document_id": documentID})
	}
}

func DeleteDocument(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		documentID := c.Param("id")

		_, err := db.Exec(context.Background(),
			`DELETE FROM documents WHERE id = $1 AND tenant_id = $2`, documentID, tenantID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete document"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}
