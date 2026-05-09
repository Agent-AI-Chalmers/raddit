import React, { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client'

const SUBREDDITS = ['general', 'technology', 'gaming', 'science', 'news', 'funny']

export default function CreatePost() {
  const navigate = useNavigate()
  const [form, setForm] = useState({ title: '', content: '', subreddit: 'general' })
  const [urlToPreview, setUrlToPreview] = useState('')
  const [preview, setPreview] = useState(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const data = await api.createPost(form)
      navigate(`/posts/${data.post_id}`)
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  const handlePreview = async () => {
    if (!urlToPreview) return
    try {
      // Defense-in-depth: validate URL scheme before sending to server
      const parsed = new URL(urlToPreview)
      if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
        setError('Only http and https URLs are allowed')
        return
      }
      const data = await api.previewURL(urlToPreview)
      setPreview(data)
    } catch (e) {
      setError('Could not fetch URL preview')
    }
  }

  return (
    <div className="container">
      <div className="main-content">
        <div className="card" style={{ padding: 24 }}>
          <h2 style={{ marginBottom: 20 }}>Create a Post</h2>

          {error && <div className="error-msg">{error}</div>}

          <form onSubmit={handleSubmit}>
            <div style={{ marginBottom: 16 }}>
              <label style={{ display: 'block', fontWeight: 700, marginBottom: 6, fontSize: 13 }}>Community</label>
              <select
                value={form.subreddit}
                onChange={e => setForm({ ...form, subreddit: e.target.value })}
                style={{ width: 'auto', minWidth: 200 }}
              >
                {SUBREDDITS.map(s => <option key={s} value={s}>r/{s}</option>)}
              </select>
            </div>

            <div style={{ marginBottom: 16 }}>
              <label style={{ display: 'block', fontWeight: 700, marginBottom: 6, fontSize: 13 }}>Title</label>
              <input
                type="text"
                value={form.title}
                onChange={e => setForm({ ...form, title: e.target.value })}
                placeholder="Post title…"
                required
              />
            </div>

            <div style={{ marginBottom: 16 }}>
              <label style={{ display: 'block', fontWeight: 700, marginBottom: 6, fontSize: 13 }}>
                Content <span style={{ color: '#878a8c', fontWeight: 400 }}>(HTML supported)</span>
              </label>
              <textarea
                value={form.content}
                onChange={e => setForm({ ...form, content: e.target.value })}
                placeholder="What are your thoughts? HTML formatting is supported…"
                style={{ minHeight: 200 }}
                required
              />
            </div>

            {/* Link post preview utility */}
            <div style={{ marginBottom: 20, padding: 12, background: '#f8f9fa', borderRadius: 4 }}>
              <label style={{ display: 'block', fontWeight: 700, marginBottom: 6, fontSize: 13 }}>
                Link Preview (optional)
              </label>
              <div style={{ display: 'flex', gap: 8 }}>
                <input
                  type="text"
                  value={urlToPreview}
                  onChange={e => setUrlToPreview(e.target.value)}
                  placeholder="https://example.com"
                />
                <button type="button" className="btn-outline btn-sm" onClick={handlePreview} style={{ whiteSpace: 'nowrap' }}>
                  Fetch Preview
                </button>
              </div>
              {preview && (
                <div style={{ marginTop: 8, fontSize: 12, color: '#878a8c' }}>
                  Status: {preview.status_code} · Type: {preview.content_type}
                </div>
              )}
            </div>

            <div style={{ display: 'flex', gap: 8 }}>
              <button type="submit" className="btn-primary" disabled={loading}>
                {loading ? 'Posting…' : 'Post'}
              </button>
              <button type="button" className="btn-outline" onClick={() => navigate(-1)}>
                Cancel
              </button>
            </div>
          </form>
        </div>
      </div>

      <aside className="sidebar">
        <div className="card" style={{ padding: 16 }}>
          <h4 style={{ marginBottom: 8 }}>Posting Tips</h4>
          <ul style={{ fontSize: 13, color: '#3c3c3c', paddingLeft: 16 }}>
            <li style={{ marginBottom: 6 }}>Be respectful and constructive</li>
            <li style={{ marginBottom: 6 }}>HTML tags are supported in post content</li>
            <li style={{ marginBottom: 6 }}>Use the link preview to embed references</li>
          </ul>
        </div>
      </aside>
    </div>
  )
}
