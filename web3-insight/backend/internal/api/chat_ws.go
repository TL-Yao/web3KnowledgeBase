package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/user/web3-insight/internal/service"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in dev
	},
}

// ChatRequest represents a chat message from the client
type ChatRequest struct {
	ArticleID    string `json:"articleId"`
	Message      string `json:"message"`
	SelectedText string `json:"selectedText,omitempty"`
	SessionID    string `json:"sessionId"`
}

// ChatResponse represents a response chunk sent to the client
type ChatResponse struct {
	Type    string `json:"type"` // "chunk", "done", "error"
	Content string `json:"content,omitempty"`
	Model   string `json:"model,omitempty"`
}

// ChatHandler handles WebSocket chat connections
type ChatHandler struct {
	chatService *service.ChatService
}

// NewChatHandler creates a new chat handler
func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// HandleWebSocket handles WebSocket connections for chat
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Use mutex to prevent concurrent writes
	var writeMu sync.Mutex

	writeJSON := func(resp ChatResponse) error {
		writeMu.Lock()
		defer writeMu.Unlock()
		return conn.WriteJSON(resp)
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var req ChatRequest
		if err := json.Unmarshal(message, &req); err != nil {
			writeJSON(ChatResponse{Type: "error", Content: "Invalid request format"})
			continue
		}

		if req.Message == "" {
			writeJSON(ChatResponse{Type: "error", Content: "Message cannot be empty"})
			continue
		}

		// Stream response from LLM
		stream, model, err := h.chatService.Chat(req.ArticleID, req.Message, req.SelectedText)
		if err != nil {
			writeJSON(ChatResponse{Type: "error", Content: err.Error()})
			continue
		}

		// Send each chunk to the client
		for chunk := range stream {
			if chunk.Error != nil {
				writeJSON(ChatResponse{Type: "error", Content: chunk.Error.Error()})
				break
			}
			if chunk.Done {
				break
			}
			if chunk.Content != "" {
				if err := writeJSON(ChatResponse{Type: "chunk", Content: chunk.Content}); err != nil {
					log.Printf("Failed to write chunk: %v", err)
					break
				}
			}
		}

		writeJSON(ChatResponse{Type: "done", Model: model})
	}
}
