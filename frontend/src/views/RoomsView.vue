<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useChatStore } from '../stores/chat'

const chat = useChatStore()
const router = useRouter()

const name = ref('')
const description = ref('')
const isPrivate = ref(false)
const error = ref('')
const creating = ref(false)

onMounted(() => {
  chat.fetchRooms().catch((e) => (error.value = e.message))
})

async function createRoom() {
  if (!name.value.trim()) return
  creating.value = true
  error.value = ''
  try {
    const room = await chat.createRoom(name.value, description.value, isPrivate.value)
    name.value = ''
    description.value = ''
    isPrivate.value = false
    router.push({ name: 'room', params: { id: room.id } })
  } catch (e) {
    error.value = e.message
  } finally {
    creating.value = false
  }
}

function openRoom(id) {
  router.push({ name: 'room', params: { id } })
}
</script>

<template>
  <div class="rooms-page">
    <section class="create-room">
      <h2>Создать комнату</h2>
      <form @submit.prevent="createRoom">
        <input v-model="name" placeholder="Название" required minlength="2" />
        <input v-model="description" placeholder="Описание (необязательно)" />
        <label class="checkbox">
          <input v-model="isPrivate" type="checkbox" />
          Приватная
        </label>
        <button type="submit" :disabled="creating">Создать</button>
      </form>
      <p v-if="error" class="error">{{ error }}</p>
    </section>

    <section class="room-list">
      <h2>Комнаты</h2>
      <ul>
        <li v-for="room in chat.rooms" :key="room.id" @click="openRoom(room.id)">
          <div class="room-name">
            {{ room.name }}
            <span v-if="room.is_private" class="badge">приватная</span>
          </div>
          <div class="room-desc">{{ room.description }}</div>
        </li>
      </ul>
      <p v-if="!chat.rooms.length">Комнат пока нет — создайте первую.</p>
    </section>
  </div>
</template>

<style scoped>
.rooms-page {
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 24px;
  padding: 24px;
}
.create-room, .room-list {
  background: #1e293b;
  border-radius: 12px;
  padding: 20px;
}
form {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
}
ul {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
li {
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 12px 14px;
  cursor: pointer;
}
li:hover {
  border-color: #38bdf8;
}
.room-name {
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}
.badge {
  font-size: 0.7rem;
  background: #334155;
  padding: 2px 8px;
  border-radius: 999px;
  font-weight: 400;
}
.room-desc {
  color: #94a3b8;
  font-size: 0.85rem;
  margin-top: 4px;
}
.error {
  color: #f87171;
  font-size: 0.85rem;
}
@media (max-width: 720px) {
  .rooms-page {
    grid-template-columns: 1fr;
  }
}
</style>
