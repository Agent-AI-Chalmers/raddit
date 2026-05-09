import React, { createContext, useContext, useState, useEffect } from 'react'

const AuthContext = createContext(null)

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null)
  const [loading, setLoading] = useState(true)

  // Session is persisted via HttpOnly cookie set by the backend;
  // no token is stored in localStorage to prevent XSS exfiltration.
  useEffect(() => {
    fetch('/api/auth/me', { credentials: 'include' })
      .then(r => r.ok ? r.json() : Promise.reject())
      .then(data => setUser(data))
      .catch(() => setUser(null))
      .finally(() => setLoading(false))
  }, [])

  const login = (userData) => {
    // The HttpOnly session cookie is already set by the backend response.
    // Do not store the JWT in localStorage or JavaScript-accessible state.
    setUser(userData)
  }

  const logout = () => {
    // Clear the server-side session cookie
    fetch('/api/auth/logout', { method: 'POST', credentials: 'include' }).catch(() => {})
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, login, logout, loading }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => useContext(AuthContext)
