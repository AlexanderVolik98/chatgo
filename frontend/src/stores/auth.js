import { defineStore } from 'pinia'
import { api } from '../api/client'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('chatgo_token') || null,
    user: JSON.parse(localStorage.getItem('chatgo_user') || 'null'),
  }),

  getters: {
    isAuthenticated: (state) => !!state.token,
  },

  actions: {
    setUser(user) {
      this.user = user
      localStorage.setItem('chatgo_user', JSON.stringify(user))
    },

    setSession(token, user) {
      this.token = token
      localStorage.setItem('chatgo_token', token)
      this.setUser(user)
    },

    async refreshUser() {
      this.setUser(await api.me(this.token))
    },

    async updateUsername(name) {
      const data = await api.updateMe(this.token, name)
      this.setUser(data.user)
    },

    async register(username, email, password) {
      const data = await api.register(username, email, password)
      this.setSession(data.token, data.user)
    },

    async login(email, password) {
      const data = await api.login(email, password)
      this.setSession(data.token, data.user)
    },

    logout() {
      this.token = null
      this.user = null
      localStorage.removeItem('chatgo_token')
      localStorage.removeItem('chatgo_user')
    },
  },
})
