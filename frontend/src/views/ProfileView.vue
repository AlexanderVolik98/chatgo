<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useChatStore } from '../stores/chat'

const auth = useAuthStore()
const chat = useChatStore()

const name = ref(auth.user?.username || '')
const error = ref('')
const saved = ref(false)
const saving = ref(false)

const trimmed = computed(() => name.value.trim())
const canSave = computed(
  () => !saving.value && trimmed.value.length >= 2 && trimmed.value !== auth.user?.username,
)
const memberSince = computed(() =>
  auth.user?.created_at ? new Date(auth.user.created_at).toLocaleDateString('ru-RU') : '',
)

onMounted(async () => {
  try {
    await auth.refreshUser()
    name.value = auth.user.username
  } catch (e) {
    error.value = e.message
  }
})

function onInput() {
  saved.value = false
  error.value = ''
}

async function save() {
  if (!canSave.value) return
  error.value = ''
  saved.value = false
  saving.value = true
  try {
    await auth.updateUsername(trimmed.value)
    name.value = auth.user.username
    saved.value = true
    // The socket keeps the username from its first join; reconnect so new messages use the new name.
    chat.disconnect()
  } catch (e) {
    error.value = e.status === 409 ? 'Это имя уже занято' : e.message
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="profile-page">
    <section class="card">
      <h2>Профиль</h2>

      <dl class="info">
        <dt>Email</dt>
        <dd>{{ auth.user?.email }}</dd>
        <dt>В чате с</dt>
        <dd>{{ memberSince }}</dd>
      </dl>

      <form @submit.prevent="save">
        <label>
          Имя пользователя
          <input v-model="name" minlength="2" required autocomplete="off" @input="onInput" />
        </label>
        <p v-if="error" class="error">{{ error }}</p>
        <p v-else-if="saved" class="success">Сохранено</p>
        <button type="submit" :disabled="!canSave">{{ saving ? 'Сохраняем...' : 'Сохранить' }}</button>
      </form>
    </section>
  </div>
</template>

<style scoped>
.profile-page {
  display: flex;
  justify-content: center;
  padding: 24px;
}
.card {
  background: #1e293b;
  border-radius: 12px;
  padding: 24px;
  width: 100%;
  max-width: 420px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
h2 {
  margin: 0;
}
.info {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 8px 16px;
  margin: 0;
}
dt {
  color: #94a3b8;
}
dd {
  margin: 0;
  word-break: break-all;
}
form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
label {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 0.9rem;
}
button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
.error {
  color: #f87171;
  font-size: 0.85rem;
  margin: 0;
}
.success {
  color: #4ade80;
  font-size: 0.85rem;
  margin: 0;
}
</style>
