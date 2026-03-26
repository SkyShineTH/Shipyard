import { useCallback, useMemo, useState, useEffect } from 'react'
import { AuthContext } from './authContext'

const STORAGE_KEY = 'shipyard_token'

export function AuthProvider({ children }) {
  const [token, setTokenState] = useState(
    () => localStorage.getItem(STORAGE_KEY) || null,
  )

  useEffect(() => {
    if (token) localStorage.setItem(STORAGE_KEY, token)
    else localStorage.removeItem(STORAGE_KEY)
  }, [token])

  const setToken = useCallback((t) => setTokenState(t), [])
  const logout = useCallback(() => setTokenState(null), [])

  const value = useMemo(
    () => ({
      token,
      setToken,
      logout,
      isAuthenticated: Boolean(token),
    }),
    [token, setToken, logout],
  )

  return (
    <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
  )
}
