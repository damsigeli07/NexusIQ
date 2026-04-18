package handlers

import (
	"bytes"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/damsigeli07/NexusIQ/internal/config"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // tighten this in production
	},
}

type wsMessage struct {
	Question string `json:"question"`
}

type mlStreamRequest struct {
	Question string `json:"question"`
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
}

func ChatWebSocket(db *pgxpool.Pool, rdb *redis.Client, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetString("tenant_id")
		userID   := c.GetString("user_id")

		// Upgrade HTTP → WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			// Read question from client
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}

			var incoming wsMessage
			if err := json.Unmarshal(msg, &incoming); err != nil || incoming.Question == "" {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"invalid message"}`))
				continue
			}

			// Check Redis cache first
			cacheKey := fmt.Sprintf("chat:%s:%s", tenantID, incoming.Question)
			cached, err := rdb.Get(context.Background(), cacheKey).Result()
			if err == nil {
				// Cache hit — send full answer immediately
				conn.WriteMessage(websocket.TextMessage,
					[]byte(fmt.Sprintf(`{"type":"cached","content":%q}`, cached)))
				continue
			}

			// Stream from Python ML service
			fullAnswer, err := streamFromML(conn, cfg, mlStreamRequest{
				Question: incoming.Question,
				TenantID: tenantID,
				UserID:   userID,
			})
			if err != nil {
				conn.WriteMessage(websocket.TextMessage,
					[]byte(fmt.Sprintf(`{"type":"error","content":%q}`, err.Error())))
				continue
			}

			// Cache the answer for 1 hour
			rdb.Set(context.Background(), cacheKey, fullAnswer, time.Hour)

			// Persist to chat_history
			go saveChatHistory(db, tenantID, userID, incoming.Question, fullAnswer)

			// Signal stream complete
			conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"done"}`))
		}
	}
}

// streamFromML calls Python /stream endpoint and forwards each SSE token to the WebSocket client
func streamFromML(conn *websocket.Conn, cfg *config.Config, req mlStreamRequest) (string, error) {
	mlURL := os.Getenv("ML_SERVICE_URL")
	if mlURL == "" {
		mlURL = "http://python-ml:8000"
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(mlURL+"/stream", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ml service unavailable: %w", err)
	}
	defer resp.Body.Close()

	var fullAnswer string
	scanner := bufio.NewScanner(resp.Body)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		// SSE format: "data: <token>"
		if len(line) > 6 && line[:5] == "data:" {
			token := line[6:] // strip "data: "

			if token == "[DONE]" {
				break
			}

			fullAnswer += token

			// Forward token to WebSocket client
			wsPayload := fmt.Sprintf(`{"type":"token","content":%q}`, token)
			if err := conn.WriteMessage(websocket.TextMessage, []byte(wsPayload)); err != nil {
				return fullAnswer, nil // client disconnected — that's ok
			}
		}
	}

	return fullAnswer, nil
}

func saveChatHistory(db *pgxpool.Pool, tenantID, userID, question, answer string) {
	db.Exec(context.Background(),
		`INSERT INTO chat_history (tenant_id, user_id, question, answer)
		 VALUES ($1, $2, $3, $4)`,
		tenantID, userID, question, answer,
	)
}
