package api

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/user/web3-insight/internal/llm"
	"github.com/user/web3-insight/internal/service"
)

// ResearchChatRequest represents a research chat message from the client.
type ResearchChatRequest struct {
	SessionID string        `json:"sessionId"`
	Message   string        `json:"message"`
	History   []ChatMessage `json:"history,omitempty"`
	Model     string        `json:"model,omitempty"`
}

// ResearchChatHandler handles WebSocket chat for research sessions.
type ResearchChatHandler struct {
	chatService *service.ResearchChatService
}

func NewResearchChatHandler(chatService *service.ResearchChatService) *ResearchChatHandler {
	return &ResearchChatHandler{
		chatService: chatService,
	}
}

// HandleWebSocket handles WebSocket connections for research chat.
func (h *ResearchChatHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade research chat connection: %v", err)
		return
	}
	defer conn.Close()

	conn.SetReadLimit(32 * 1024)

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
				log.Printf("Research WebSocket error: %v", err)
			}
			break
		}

		var req ResearchChatRequest
		if err := json.Unmarshal(message, &req); err != nil {
			writeJSON(ChatResponse{Type: "error", Content: "Invalid request format"})
			continue
		}

		if req.Message == "" {
			writeJSON(ChatResponse{Type: "error", Content: "Message cannot be empty"})
			continue
		}

		var stream <-chan llm.StreamChunk
		var model string

		if len(req.History) > 0 {
			llmMessages := make([]llm.Message, 0, len(req.History)+1)
			for _, m := range req.History {
				llmMessages = append(llmMessages, llm.Message{Role: m.Role, Content: m.Content})
			}
			llmMessages = append(llmMessages, llm.Message{Role: "user", Content: req.Message})
			stream, model, err = h.chatService.ChatWithMessagesAndModel(req.SessionID, llmMessages, req.Model)
		} else {
			stream, model, err = h.chatService.ChatWithModel(req.SessionID, req.Message, req.Model)
		}

		if err != nil {
			writeJSON(ChatResponse{Type: "error", Content: fmt.Sprintf("Chat error: %v", err)})
			continue
		}

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
					log.Printf("Failed to write research chat chunk: %v", err)
					break
				}
			}
		}

		writeJSON(ChatResponse{Type: "done", Model: model})
	}
}
