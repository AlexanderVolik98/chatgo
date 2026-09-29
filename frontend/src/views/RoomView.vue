<script setup>
import { ref, computed, watch, onUnmounted, nextTick } from 'vue'
import { useChatStore } from '../stores/chat'
import { useAuthStore } from '../stores/auth'

const props = defineProps({ id: { type: [String, Number], required: true } })

const chat = useChatStore()
const auth = useAuthStore()

const messageText = ref('')
const error = ref('')
const joining = ref(false)
const scrollBox = ref(null)

const roomId = computed(() => Number(props.id))

async function load() {
  error.value = ''
  try {
    await chat.openRoom(roomId.value)
    await nextTick()
    scrollToBottom()
  } catch (e) {
    error.value = e.message
  }
}

watch(roomId, load, { immediate: true })

watch(
  () => chat.messages.length,
  async () => {
    await nextTick()
    scrollToBottom()
  }
)

function scrollToBottom() {
  if (scrollBox.value) {
    scrollBox.value.scrollTop = scrollBox.value.scrollHeight
  }
}

onUnmounted(() => {
  chat.closeRoom()
})

function submit() {
  if (!messageText.value.trim()) return
  chat.sendMessage(messageText.value.trim())
  messageText.value = ''
}

async function join() {
  joining.value = true
  error.value = ''
  try {
    await chat.joinRoom(roomId.value)
    await chat.openRoom(roomId.value)
  } catch (e) {
    error.value = e.message
  } finally {
    joining.value = false
  }
}

async function leave() {
  try {
    await chat.leaveRoom(roomId.value)
    await chat.openRoom(roomId.value)
  } catch (e) {
    error.value = e.message
  }
}

async function toggleAdmin(member) {
  const newRole = member.role === 'admin' ? 'member' : 'admin'
  try {
    await chat.setMemberRole(member.user_id, newRole)
  } catch (e) {
    error.value = e.message
  }
}

async function kick(member) {
  try {
    await chat.removeMember(member.user_id)
  } catch (e) {
    error.value = e.message
  }
}

async function removeMessage(messageId) {
  try {
    await chat.deleteMessage(messageId)
  } catch (e) {
    error.value = e.message
  }
}

function canDelete(message) {
  return message.user_id === auth.user?.id || chat.isCurrentUserAdmin
}

function formatTime(iso) {
  return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
</script>

<template>
  <div class="room-page" v-if="chat.currentRoom">
    <section class="chat-column">
      <header class="room-header">
        <div>
          <h2>{{ chat.currentRoom.name }}</h2>
          <p class="desc">{{ chat.currentRoom.description }}</p>
        </div>
        <button v-if="chat.isCurrentUserMember && chat.currentUserRole !== 'owner'" @click="leave">Покинуть</button>
      </header>

      <p v-if="error" class="error">{{ error }}</p>

      <div v-if="!chat.isCurrentUserMember" class="join-box">
        <p>Вы не участник этой комнаты.</p>
        <button @click="join" :disabled="joining">{{ joining ? 'Вступаем...' : 'Вступить' }}</button>
      </div>

      <template v-else>
        <div class="messages" ref="scrollBox">
          <div v-for="m in chat.messages" :key="m.id" class="message">
            <div class="message-meta">
              <span class="author">{{ m.username }}</span>
              <span class="time">{{ formatTime(m.created_at) }}</span>
              <button v-if="canDelete(m)" class="delete-btn" @click="removeMessage(m.id)">✕</button>
            </div>
            <div class="message-content">{{ m.content }}</div>
          </div>
          <p v-if="!chat.messages.length" class="empty">Сообщений пока нет.</p>
        </div>

        <form class="composer" @submit.prevent="submit">
          <input v-model="messageText" placeholder="Написать сообщение..." autocomplete="off" />
          <button type="submit">Отправить</button>
        </form>
      </template>
    </section>

    <aside class="members-column">
      <h3>Участники ({{ chat.members.length }})</h3>
      <ul>
        <li v-for="m in chat.members" :key="m.user_id">
          <div class="member-row">
            <span>{{ m.username }}</span>
            <span class="role" :class="m.role">{{ m.role }}</span>
          </div>
          <div v-if="chat.isCurrentUserAdmin && m.role !== 'owner'" class="member-actions">
            <button @click="toggleAdmin(m)">{{ m.role === 'admin' ? 'Разжаловать' : 'Сделать админом' }}</button>
            <button @click="kick(m)">Удалить</button>
          </div>
        </li>
      </ul>
    </aside>
  </div>
</template>

<style scoped>
.room-page {
  display: grid;
  grid-template-columns: 1fr 260px;
  gap: 20px;
  padding: 20px;
  flex: 1;
  min-height: 0;
}
.chat-column {
  display: flex;
  flex-direction: column;
  background: #1e293b;
  border-radius: 12px;
  padding: 16px;
  min-height: 0;
}
.room-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  border-bottom: 1px solid #334155;
  padding-bottom: 12px;
  margin-bottom: 12px;
}
.desc {
  color: #94a3b8;
  margin: 4px 0 0;
  font-size: 0.85rem;
}
.messages {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-right: 6px;
}
.message-meta {
  display: flex;
  gap: 8px;
  align-items: center;
  font-size: 0.75rem;
  color: #94a3b8;
}
.author {
  font-weight: 600;
  color: #38bdf8;
}
.delete-btn {
  background: transparent;
  color: #f87171;
  padding: 0 4px;
  font-weight: 700;
}
.delete-btn:hover {
  background: transparent;
  color: #ef4444;
}
.message-content {
  margin-top: 2px;
}
.composer {
  display: flex;
  gap: 10px;
  margin-top: 12px;
}
.composer input {
  flex: 1;
}
.join-box {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  padding: 20px 0;
}
.members-column {
  background: #1e293b;
  border-radius: 12px;
  padding: 16px;
}
.members-column ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.member-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.role {
  font-size: 0.7rem;
  padding: 2px 8px;
  border-radius: 999px;
  background: #334155;
}
.role.owner {
  background: #f59e0b;
  color: #0f172a;
}
.role.admin {
  background: #38bdf8;
  color: #0f172a;
}
.member-actions {
  display: flex;
  gap: 6px;
  margin-top: 6px;
}
.member-actions button {
  font-size: 0.75rem;
  padding: 4px 8px;
  background: #334155;
  color: #e2e8f0;
}
.error {
  color: #f87171;
  font-size: 0.85rem;
}
.empty {
  color: #64748b;
  font-size: 0.85rem;
}
@media (max-width: 800px) {
  .room-page {
    grid-template-columns: 1fr;
  }
}
</style>
