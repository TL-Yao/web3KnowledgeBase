package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/service"
)

var wsAllowedOrigins = map[string]bool{
	"http://localhost:3000":  true,
	"http://127.0.0.1:3000": true,
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return wsAllowedOrigins[origin]
	},
}

// ChatMessage represents a single message in conversation history
type ChatMessage struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"`
}

// ChatRequest represents a chat message from the client
type ChatRequest struct {
	ArticleID    string        `json:"articleId"`
	Message      string        `json:"message"`
	SelectedText string        `json:"selectedText,omitempty"`
	SessionID    string        `json:"sessionId"`
	History      []ChatMessage `json:"history,omitempty"`
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

	// Limit max message size to 32KB
	conn.SetReadLimit(32 * 1024)

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
		var stream <-chan llm.StreamChunk
		var model string

		if len(req.History) > 0 {
			// Multi-turn: convert history + current message to llm.Message slice
			llmMessages := make([]llm.Message, 0, len(req.History)+1)
			for _, m := range req.History {
				llmMessages = append(llmMessages, llm.Message{Role: m.Role, Content: m.Content})
			}
			userMsg := req.Message
			if req.SelectedText != "" {
				userMsg = fmt.Sprintf("关于「%s」这部分内容：%s", req.SelectedText, req.Message)
			}
			llmMessages = append(llmMessages, llm.Message{Role: "user", Content: userMsg})
			stream, model, err = h.chatService.ChatWithMessages(req.ArticleID, llmMessages)
		} else {
			stream, model, err = h.chatService.Chat(req.ArticleID, req.Message, req.SelectedText)
		}
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
