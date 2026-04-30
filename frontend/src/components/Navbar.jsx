import React, { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { api } from '../api/client'

export default function Navbar() {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [searchQ, setSearchQ] = useState('')

  const handleSearch = (e) => {
    e.preventDefault()
    if (searchQ.trim()) {
      navigate(`/?q=${encodeURIComponent(searchQ)}`)
    }
  }

  const handleLogout = async () => {
    await api.logout()
    logout()
    navigate('/')
  }

  return (
    <nav style={{
      background: '#fff',
      borderBottom: '1px solid #edeff1',
      padding: '0 20px',
      height: 48,
      display: 'flex',
      alignItems: 'center',
      gap: 16,
      position: 'sticky',
      top: 0,
      zIndex: 100,
    }}>
      <Link to="/" style={{ fontWeight: 800, fontSize: 18, color: '#ff4500', textDecoration: 'none' }}>
        raddit
      </Link>

      <form onSubmit={handleSearch} style={{ flex: 1, maxWidth: 690, display: 'flex', gap: 8 }}>
        <input
          type="text"
          placeholder="Search Raddit…"
          value={searchQ}
          onChange={e => setSearchQ(e.target.value)}
          style={{ width: '100%' }}
        />
        <button type="submit" className="btn-primary btn-sm">Search</button>
      </form>

      <div style={{ display: 'flex', gap: 8, alignItems: 'center', marginLeft: 'auto' }}>
        <Link to="/tools" style={{ fontSize: 13, color: '#878a8c' }}>Tools</Link>
        {user ? (
          <>
            <Link to="/submit"><button className="btn-orange btn-sm">+ New Post</button></Link>
            <Link to="/profile" style={{ fontSize: 13 }}>u/{user.username}</Link>
            {user.role === 'admin' && <Link to="/admin" style={{ fontSize: 13, color: '#ff4500' }}>Admin</Link>}
            <button className="btn-outline btn-sm" onClick={handleLogout}>Logout</button>
          </>
        ) : (
          <>
            <Link to="/login"><button className="btn-outline btn-sm">Log In</button></Link>
            <Link to="/register"><button className="btn-orange btn-sm">Sign Up</button></Link>
          </>
        )}
      </div>
    </nav>
  )
}
