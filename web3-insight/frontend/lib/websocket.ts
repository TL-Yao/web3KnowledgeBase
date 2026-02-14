const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'

export interface ChatMessage {
  type: 'chunk' | 'done' | 'error'
  content?: string
  model?: string
}

export function createChatWebSocket(onMessage: (data: ChatMessage) => void) {
  const ws = new WebSocket(`${WS_URL}/ws/chat`)

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data) as ChatMessage
      onMessage(data)
    } catch {
      console.error('Failed to parse WebSocket message:', event.data)
    }
  }

  ws.onerror = () => {
    // Silently handle connection failures — backend may not be running
  }

  return ws
}

export function createResearchChatWebSocket(onMessage: (data: ChatMessage) => void) {
  const ws = new WebSocket(`${WS_URL}/ws/research-chat`)

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data) as ChatMessage
      onMessage(data)
    } catch {
      console.error('Failed to parse WebSocket message:', event.data)
    }
  }

  ws.onerror = () => {
    // Silently handle connection failures — backend may not be running
  }

  return ws
}
