<script setup>
import { useAuthStore } from './stores/auth'
import { useChatStore } from './stores/chat'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const chat = useChatStore()
const router = useRouter()

function logout() {
  chat.disconnect()
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <div class="app">
    <header class="topbar">
      <router-link to="/rooms" class="brand">ChatGo</router-link>
      <div v-if="auth.isAuthenticated" class="user-box">
        <span>{{ auth.user?.username }}</span>
        <button @click="logout">Выйти</button>
      </div>
    </header>
    <main class="content">
      <router-view />
    </main>
  </div>
</template>

<style>
* {
  box-sizing: border-box;
}
body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #0f172a;
  color: #e2e8f0;
}
.app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 20px;
  background: #1e293b;
  border-bottom: 1px solid #334155;
}
.brand {
  font-weight: 700;
  font-size: 1.2rem;
  color: #38bdf8;
  text-decoration: none;
}
.user-box {
  display: flex;
  align-items: center;
  gap: 12px;
}
.content {
  flex: 1;
  display: flex;
  flex-direction: column;
}
button {
  cursor: pointer;
  border: none;
  border-radius: 6px;
  padding: 8px 14px;
  background: #38bdf8;
  color: #0f172a;
  font-weight: 600;
}
button:hover {
  background: #0ea5e9;
}
input, textarea {
  border-radius: 6px;
  border: 1px solid #334155;
  background: #0f172a;
  color: #e2e8f0;
  padding: 8px 10px;
}
</style>
