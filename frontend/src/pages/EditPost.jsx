import React, { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../api/client'

export default function EditPost() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [form, setForm] = useState({ title: '', content: '' })
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    api.getPost(id)
      .then(data => setForm({ title: data.title, content: data.content }))
      .catch(e => setError(e.message))
  }, [id])

  const handleSubmit = async (e) => {
    e.preventDefault()
    setLoading(true)
    try {
      await api.updatePost(id, form)
      navigate(`/posts/${id}`)
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="container">
      <div className="main-content">
        <div className="card" style={{ padding: 24 }}>
          <h2 style={{ marginBottom: 20 }}>Edit Post</h2>
          {error && <div className="error-msg">{error}</div>}
          <form onSubmit={handleSubmit}>
            <div style={{ marginBottom: 16 }}>
              <label style={{ display: 'block', fontWeight: 700, marginBottom: 6, fontSize: 13 }}>Title</label>
              <input
                type="text"
                value={form.title}
                onChange={e => setForm({ ...form, title: e.target.value })}
                required
              />
            </div>
            <div style={{ marginBottom: 20 }}>
              <label style={{ display: 'block', fontWeight: 700, marginBottom: 6, fontSize: 13 }}>Content</label>
              <textarea
                value={form.content}
                onChange={e => setForm({ ...form, content: e.target.value })}
                style={{ minHeight: 200 }}
                required
              />
            </div>
            <div style={{ display: 'flex', gap: 8 }}>
              <button type="submit" className="btn-primary" disabled={loading}>
                {loading ? 'Saving…' : 'Save Changes'}
              </button>
              <button type="button" className="btn-outline" onClick={() => navigate(`/posts/${id}`)}>
                Cancel
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}
