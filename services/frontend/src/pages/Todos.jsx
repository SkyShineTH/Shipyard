import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  createTodo,
  deleteTodo,
  getTodos,
  updateTodo,
} from '../api/client'
import { useAuth } from '../context/authContext'

export default function Todos() {
  const { token, isAuthenticated } = useAuth()
  const [todos, setTodos] = useState([])
  const [title, setTitle] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)

  const load = useCallback(async () => {
    if (!token) {
      setTodos([])
      setLoading(false)
      return
    }
    setError('')
    setLoading(true)
    try {
      const list = await getTodos(token)
      setTodos(Array.isArray(list) ? list : [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load todos')
      setTodos([])
    } finally {
      setLoading(false)
    }
  }, [token])

  useEffect(() => {
    load()
  }, [load])

  async function handleAdd(e) {
    e.preventDefault()
    if (!token) return
    const t = title.trim()
    if (!t) return
    setSaving(true)
    setError('')
    try {
      const todo = await createTodo(t, token)
      setTodos((prev) => [...prev, todo])
      setTitle('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not add todo')
    } finally {
      setSaving(false)
    }
  }

  async function toggleCompleted(todo) {
    if (!token) return
    setError('')
    try {
      const updated = await updateTodo(
        todo.id,
        { completed: !todo.completed },
        token,
      )
      setTodos((prev) =>
        prev.map((x) => (x.id === updated.id ? updated : x)),
      )
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Update failed')
    }
  }

  async function handleDelete(id) {
    if (!token) return
    setError('')
    try {
      await deleteTodo(id, token)
      setTodos((prev) => prev.filter((x) => x.id !== id))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Delete failed')
    }
  }

  if (!isAuthenticated || !token) {
    return (
      <main className="page page--workspace todos-page">
        <p className="page-eyebrow">Tasks</p>
        <h1>Todos</h1>
        <p className="muted">
          Sign in to load your list. Tasks are stored per account in PostgreSQL
          (JWT required by the API).
        </p>
        <div className="todos-gate-actions">
          <Link to="/login" className="btn primary">
            Sign in
          </Link>
          <Link to="/register" className="btn secondary">
            Create account
          </Link>
        </div>
      </main>
    )
  }

  return (
    <main className="page page--workspace todos-page">
      <p className="page-eyebrow">Tasks</p>
      <h1>Todos</h1>
      <p className="muted">
        Private to your account — other users do not see your items.
      </p>

      <div className="card card--panel todos-panel">
        <form className="row todos-add" onSubmit={handleAdd}>
          <input
            type="text"
            placeholder="New task…"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            aria-label="New todo title"
          />
          <button type="submit" className="btn primary" disabled={saving}>
            Add
          </button>
        </form>

        {error ? <p className="alert">{error}</p> : null}

        {loading ? (
          <p className="muted todos-status">Loading…</p>
        ) : todos.length === 0 ? (
          <p className="muted todos-status">No todos yet. Add one above.</p>
        ) : (
          <ul className="todo-list">
          {todos.map((todo) => (
            <li key={todo.id} className="todo-item">
              <label className="todo-label">
                <input
                  type="checkbox"
                  checked={Boolean(todo.completed)}
                  onChange={() => toggleCompleted(todo)}
                />
                <span
                  className={
                    todo.completed ? 'todo-title done' : 'todo-title'
                  }
                >
                  {todo.title}
                </span>
              </label>
              <button
                type="button"
                className="btn ghost danger"
                onClick={() => handleDelete(todo.id)}
              >
                Delete
              </button>
            </li>
          ))}
          </ul>
        )}
      </div>
    </main>
  )
}
