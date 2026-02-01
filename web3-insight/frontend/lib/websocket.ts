export function createChatWebSocket(onMessage: (data: any) => void) {
  const ws = new WebSocket('ws://localhost:8080/ws/chat')

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data)
    onMessage(data)
  }

  return ws
}
