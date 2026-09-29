<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

const username = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.register(username.value, email.value, password.value)
    router.push({ name: 'rooms' })
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <form class="auth-card" @submit.prevent="submit">
      <h1>Регистрация</h1>
      <label>Имя пользователя<input v-model="username" required minlength="3" /></label>
      <label>Email<input v-model="email" type="email" required /></label>
      <label>Пароль<input v-model="password" type="password" required minlength="8" /></label>
      <p v-if="error" class="error">{{ error }}</p>
      <button type="submit" :disabled="loading">{{ loading ? 'Создаём...' : 'Зарегистрироваться' }}</button>
      <p>Уже есть аккаунт? <router-link to="/login">Войти</router-link></p>
    </form>
  </div>
</template>

<style scoped>
.auth-page {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}
.auth-card {
  background: #1e293b;
  padding: 32px;
  border-radius: 12px;
  width: 320px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}
label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 0.9rem;
}
.error {
  color: #f87171;
  font-size: 0.85rem;
}
</style>
