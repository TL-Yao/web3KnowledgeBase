const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'

export function createChatWebSocket(onMessage: (data: any) => void) {
  const ws = new WebSocket(`${WS_URL}/ws/chat`)

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data)
    onMessage(data)
  }

  return ws
}
