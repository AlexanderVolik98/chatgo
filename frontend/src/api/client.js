const API_URL = import.meta.env.VITE_API_URL || ''

class ApiError extends Error {
  constructor(message, status) {
    super(message)
    this.status = status
  }
}

async function request(path, { method = 'GET', body, token } = {}) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`

  const res = await fetch(`${API_URL}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })

  const isJson = res.headers.get('content-type')?.includes('application/json')
  const data = isJson ? await res.json().catch(() => null) : null

  if (!res.ok) {
    throw new ApiError(data?.error || `request failed with status ${res.status}`, res.status)
  }
  return data
}

export const api = {
  register: (username, email, password) =>
    request('/api/auth/register', { method: 'POST', body: { username, email, password } }),
  login: (email, password) =>
    request('/api/auth/login', { method: 'POST', body: { email, password } }),
  me: (token) => request('/api/me', { token }),

  listRooms: (token) => request('/api/rooms', { token }),
  createRoom: (token, name, description, isPrivate) =>
    request('/api/rooms', { method: 'POST', token, body: { name, description, is_private: isPrivate } }),
  getRoom: (token, id) => request(`/api/rooms/${id}`, { token }),
  updateRoom: (token, id, name, description) =>
    request(`/api/rooms/${id}`, { method: 'PATCH', token, body: { name, description } }),
  joinRoom: (token, id) => request(`/api/rooms/${id}/join`, { method: 'POST', token }),
  leaveRoom: (token, id) => request(`/api/rooms/${id}/leave`, { method: 'POST', token }),
  listMessages: (token, id, beforeId) =>
    request(`/api/rooms/${id}/messages${beforeId ? `?before=${beforeId}` : ''}`, { token }),
  setMemberRole: (token, roomId, userId, role) =>
    request(`/api/rooms/${roomId}/members/${userId}/role`, { method: 'POST', token, body: { role } }),
  removeMember: (token, roomId, userId) =>
    request(`/api/rooms/${roomId}/members/${userId}`, { method: 'DELETE', token }),
  deleteMessage: (token, id) => request(`/api/messages/${id}`, { method: 'DELETE', token }),
}

export { API_URL, ApiError }
