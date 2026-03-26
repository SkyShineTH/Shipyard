import { Link, NavLink, Route, Routes, Navigate } from 'react-router-dom'
import ParallaxBackground from './components/ParallaxBackground'
import { useAuth } from './context/authContext'
import Landing from './pages/Landing'
import Login from './pages/Login'
import Register from './pages/Register'
import Todos from './pages/Todos'

function Layout() {
  const { isAuthenticated, logout } = useAuth()

  return (
    <div className="app app--workspace">
      <ParallaxBackground />
      <div className="app-shell">
        <header className="topbar topbar--workspace">
          <Link to="/" className="brand">
            Shipyard
          </Link>
          <nav className="nav">
            <NavLink to="/todos" className="nav-link">
              Todos
            </NavLink>
            {isAuthenticated ? (
              <button type="button" className="btn ghost" onClick={logout}>
                Sign out
              </button>
            ) : (
              <>
                <NavLink to="/login" className="nav-link">
                  Sign in
                </NavLink>
                <NavLink to="/register" className="nav-link cta">
                  Register
                </NavLink>
              </>
            )}
          </nav>
        </header>

        <Routes>
          <Route
            path="/"
            element={
              <main className="page landing-main">
                <Landing />
              </main>
            }
          />
          <Route path="/todos" element={<Todos />} />
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>

        <footer className="footer">
          <span>Shipyard · React + Vite</span>
        </footer>
      </div>
    </div>
  )
}

export default function App() {
  return <Layout />
}
