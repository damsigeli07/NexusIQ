package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetAnalytics(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")

		// Total documents and status breakdown
		docRows, _ := db.Query(context.Background(),
			`SELECT status, COUNT(*) FROM documents
			 WHERE tenant_id = $1 GROUP BY status`, tenantID)
		defer docRows.Close()

		docStats := map[string]int{}
		for docRows.Next() {
			var status string
			var count int
			docRows.Scan(&status, &count)
			docStats[status] = count
		}

		// Total questions asked
		var totalQuestions int
		db.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM chat_history WHERE tenant_id = $1`, tenantID,
		).Scan(&totalQuestions)

		// Top 5 most asked questions
		qRows, _ := db.Query(context.Background(),
			`SELECT question, COUNT(*) as freq
			 FROM chat_history WHERE tenant_id = $1
			 GROUP BY question ORDER BY freq DESC LIMIT 5`, tenantID)
		defer qRows.Close()

		type topQ struct {
			Question  string `json:"question"`
			Frequency int    `json:"frequency"`
		}
		var topQuestions []topQ
		for qRows.Next() {
			var q topQ
			qRows.Scan(&q.Question, &q.Frequency)
			topQuestions = append(topQuestions, q)
		}
		if topQuestions == nil {
			topQuestions = []topQ{}
		}

		// Total chunks stored (= knowledge coverage)
		var totalChunks int
		db.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM chunks WHERE tenant_id = $1`, tenantID,
		).Scan(&totalChunks)

		c.JSON(http.StatusOK, gin.H{
			"documents":      docStats,
			"total_questions": totalQuestions,
			"top_questions":  topQuestions,
			"total_chunks":   totalChunks,
		})
	}
}
