import React, { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../api/client'

export default function Register() {
  const navigate = useNavigate()
  const [form, setForm] = useState({ username: '', email: '', password: '' })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      await api.register(form)
      navigate('/login')
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ maxWidth: 400, margin: '60px auto', padding: '0 16px' }}>
      <div className="card" style={{ padding: 32 }}>
        <h1 style={{ marginBottom: 24, fontSize: 20 }}>Create Account</h1>

        {error && <div className="error-msg">{error}</div>}

        <form onSubmit={handleSubmit}>
          {[
            { key: 'username', label: 'USERNAME', type: 'text' },
            { key: 'email',    label: 'EMAIL',    type: 'email' },
            { key: 'password', label: 'PASSWORD', type: 'password' },
          ].map(({ key, label, type }) => (
            <div key={key} style={{ marginBottom: 16 }}>
              <label style={{ display: 'block', fontSize: 12, fontWeight: 700, marginBottom: 4 }}>
                {label}
              </label>
              <input
                type={type}
                value={form[key]}
                onChange={e => setForm({ ...form, [key]: e.target.value })}
                required
              />
            </div>
          ))}
          <button type="submit" className="btn-orange" style={{ width: '100%' }} disabled={loading}>
            {loading ? 'Creating account…' : 'Sign Up'}
          </button>
        </form>

        <p style={{ marginTop: 16, fontSize: 14, textAlign: 'center', color: '#878a8c' }}>
          Already a Radditor? <Link to="/login">Log In</Link>
        </p>
      </div>
    </div>
  )
}
