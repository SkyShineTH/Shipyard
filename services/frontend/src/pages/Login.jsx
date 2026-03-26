import { useState, useEffect } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { login } from '../api/client'
import { useAuth } from '../context/authContext'

export default function Login() {
  const { setToken } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [notice, setNotice] = useState('')

  useEffect(() => {
    const registeredEmail = location.state?.registered
    if (typeof registeredEmail === 'string' && registeredEmail) {
      setEmail(registeredEmail)
      setNotice('Account created. Sign in with your password.')
      navigate(location.pathname, { replace: true, state: {} })
    }
  }, [location.pathname, location.state, navigate])

  async function handleSubmit(e) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const data = await login(email, password)
      setToken(data.token)
      navigate('/todos', { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <main className="page page--workspace">
      <p className="page-eyebrow">Account</p>
      <h1>Sign in</h1>
      <p className="muted">
        Sign in with your email. A JWT is stored in this browser only
        (localStorage).
      </p>

      <form className="card card--panel form" onSubmit={handleSubmit}>
        {notice ? <p className="notice">{notice}</p> : null}
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
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </label>
        <button type="submit" className="btn primary" disabled={loading}>
          {loading ? 'Signing in…' : 'Sign in'}
        </button>
      </form>

      <p className="muted">
        No account? <Link to="/register">Create one</Link>
      </p>
    </main>
  )
}
