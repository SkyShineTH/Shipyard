/** Base URL for API calls. Empty = same-origin (Vite dev proxy or nginx in Docker). */
const base = () => (import.meta.env.VITE_API_BASE || '').replace(/\/$/, '')

async function parseBody(res) {
  const text = await res.text()
  if (!text) return null
  try {
    return JSON.parse(text)
  } catch {
    return { error: text || res.statusText }
  }
}

/**
 * JSON fetch helper; throws Error with message from `{ error }` when !res.ok.
 * @param {string} path e.g. `/api/v1/todos`
 * @param {RequestInit & { token?: string }} [options]
 */
export async function apiFetch(path, options = {}) {
  const { token, headers: hdrs, ...init } = options
  const headers = { 'Content-Type': 'application/json', ...hdrs }
  if (token) headers.Authorization = `Bearer ${token}`

  const res = await fetch(`${base()}${path}`, { ...init, headers })
  if (res.status === 204) return null

  const data = await parseBody(res)
  if (!res.ok) {
    const msg =
      (data && typeof data.error === 'string' && data.error) ||
      res.statusText ||
      'request failed'
    throw new Error(msg)
  }
  return data
}

export async function register(email, password) {
  return apiFetch('/api/v1/register', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
}

export async function login(email, password) {
  return apiFetch('/api/v1/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  })
}

export async function getTodos(token) {
  return apiFetch('/api/v1/todos', { token })
}

export async function createTodo(title, token) {
  return apiFetch('/api/v1/todos', {
    method: 'POST',
    body: JSON.stringify({ title }),
    token,
  })
}

export async function updateTodo(id, patch, token) {
  return apiFetch(`/api/v1/todos/${id}`, {
    method: 'PUT',
    body: JSON.stringify(patch),
    token,
  })
}

export async function deleteTodo(id, token) {
  return apiFetch(`/api/v1/todos/${id}`, { method: 'DELETE', token })
}
