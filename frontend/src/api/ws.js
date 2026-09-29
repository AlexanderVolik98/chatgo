const defaultWsUrl = () => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${protocol}//${window.location.host}`
}

const WS_URL = import.meta.env.VITE_WS_URL || defaultWsUrl()

export class ChatSocket {
  constructor(token) {
    this.token = token
    this.socket = null
    this.listeners = new Set()
    this.reconnectDelay = 1000
    this.shouldReconnect = true
    // Messages sent before the socket is open; flushed in onopen.
    this.outbox = []
  }

  connect() {
    this.socket = new WebSocket(`${WS_URL}/ws?token=${encodeURIComponent(this.token)}`)

    this.socket.onopen = () => {
      const pending = this.outbox
      this.outbox = []
      pending.forEach((payload) => this.socket.send(JSON.stringify(payload)))
    }

    this.socket.onmessage = (event) => {
      let data
      try {
        data = JSON.parse(event.data)
      } catch {
        return
      }
      this.listeners.forEach((fn) => fn(data))
    }

    this.socket.onclose = () => {
      if (this.shouldReconnect) {
        setTimeout(() => this.connect(), this.reconnectDelay)
      }
    }
  }

  onMessage(fn) {
    this.listeners.add(fn)
    return () => this.listeners.delete(fn)
  }

  send(payload) {
    if (this.socket?.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(payload))
    } else {
      this.outbox.push(payload)
    }
  }

  join(roomId) {
    this.send({ type: 'join', room_id: roomId })
  }

  leave(roomId) {
    this.send({ type: 'leave', room_id: roomId })
  }

  sendMessage(roomId, content) {
    this.send({ type: 'message', room_id: roomId, content })
  }

  close() {
    this.shouldReconnect = false
    this.socket?.close()
  }
}
