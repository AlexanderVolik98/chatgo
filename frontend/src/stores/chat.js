import { defineStore } from 'pinia'
import { api } from '../api/client'
import { ChatSocket } from '../api/ws'
import { useAuthStore } from './auth'

export const useChatStore = defineStore('chat', {
  state: () => ({
    rooms: [],
    currentRoom: null,
    members: [],
    messages: [],
    socket: null,
    lastError: null,
  }),

  getters: {
    isCurrentUserMember: (state) => {
      const auth = useAuthStore()
      return state.members.some((m) => m.user_id === auth.user?.id)
    },
    currentUserRole: (state) => {
      const auth = useAuthStore()
      return state.members.find((m) => m.user_id === auth.user?.id)?.role || null
    },
    isCurrentUserAdmin() {
      return this.currentUserRole === 'owner' || this.currentUserRole === 'admin'
    },
  },

  actions: {
    ensureSocket() {
      if (this.socket) return this.socket
      const auth = useAuthStore()
      this.socket = new ChatSocket(auth.token)
      this.socket.connect()
      this.socket.onMessage((data) => {
        if (data.type === 'message' && this.currentRoom && data.room_id === this.currentRoom.id) {
          this.messages.push(data.message)
        }
        if (data.type === 'error') {
          this.lastError = data.error
        }
      })
      return this.socket
    },

    async fetchRooms() {
      const auth = useAuthStore()
      this.rooms = await api.listRooms(auth.token)
    },

    async createRoom(name, description, isPrivate) {
      const auth = useAuthStore()
      const room = await api.createRoom(auth.token, name, description, isPrivate)
      this.rooms.unshift(room)
      return room
    },

    async openRoom(roomId) {
      const auth = useAuthStore()
      const detail = await api.getRoom(auth.token, roomId)
      this.currentRoom = detail
      this.members = detail.members

      const isMember = detail.members.some((m) => m.user_id === auth.user?.id)
      if (isMember) {
        this.messages = await api.listMessages(auth.token, roomId)
        this.ensureSocket().join(roomId)
      } else {
        this.messages = []
      }
    },

    closeRoom() {
      if (this.currentRoom) {
        this.socket?.leave(this.currentRoom.id)
      }
      this.currentRoom = null
      this.members = []
      this.messages = []
    },

    sendMessage(content) {
      if (!this.currentRoom) return
      this.ensureSocket().sendMessage(this.currentRoom.id, content)
    },

    async joinRoom(roomId) {
      const auth = useAuthStore()
      await api.joinRoom(auth.token, roomId)
      await this.fetchRooms()
    },

    async leaveRoom(roomId) {
      const auth = useAuthStore()
      await api.leaveRoom(auth.token, roomId)
      await this.fetchRooms()
    },

    async setMemberRole(userId, role) {
      const auth = useAuthStore()
      await api.setMemberRole(auth.token, this.currentRoom.id, userId, role)
      await this.openRoom(this.currentRoom.id)
    },

    async removeMember(userId) {
      const auth = useAuthStore()
      await api.removeMember(auth.token, this.currentRoom.id, userId)
      await this.openRoom(this.currentRoom.id)
    },

    async deleteMessage(messageId) {
      const auth = useAuthStore()
      await api.deleteMessage(auth.token, messageId)
      this.messages = this.messages.filter((m) => m.id !== messageId)
    },

    disconnect() {
      this.socket?.close()
      this.socket = null
    },
  },
})
