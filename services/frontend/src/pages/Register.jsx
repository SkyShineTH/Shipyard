import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { register } from '../api/client'

export default function Register() {
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await register(email, password)
      navigate('/login', { replace: true, state: { registered: email } })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="page page--workspace">
      <p className="page-eyebrow">Account</p>
      <h1>Create account</h1>
      <p className="muted">Password must be at least 8 characters.</p>

      <form className="card card--panel form" onSubmit={handleSubmit}>
        {error ? <p className="alert">{error}</p> : null}
        <label>
          Email
          <input
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </label>
        <label>
          Password
          <input
            type="password"
            autoComplete="new-password"
            minLength={8}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </label>
        <button type="submit" className="btn primary" disabled={loading}>
          {loading ? 'Creating…' : 'Register'}
        </button>
      </form>

      <p className="muted">
        Already have an account? <Link to="/login">Sign in</Link>
      </p>
    </main>
  )
}
