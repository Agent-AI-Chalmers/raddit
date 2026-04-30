import React, { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '../api/client'
import { useAuth } from '../contexts/AuthContext'
import PostCard from '../components/PostCard'

export default function Profile() {
  const { id } = useParams()
  const { user: currentUser } = useAuth()
  const [profile, setProfile] = useState(null)
  const [posts, setPosts] = useState([])
  const [files, setFiles] = useState([])
  const [loading, setLoading] = useState(true)
  const [editMode, setEditMode] = useState(false)
  const [form, setForm] = useState({ email: '', bio: '', avatar: '', password: '', role: '' })
  const [uploadFile, setUploadFile] = useState(null)
  const [msg, setMsg] = useState('')
  const [error, setError] = useState('')

  // Determine whose profile we're viewing
  const profileId = id || currentUser?.id
  const isOwn = !id || (currentUser && parseInt(id) === currentUser.id)

  useEffect(() => {
    if (!profileId) return
    setLoading(true)
    Promise.all([
      api.getUser(profileId),
      api.getUserPosts(profileId),
      isOwn ? api.listFiles() : Promise.resolve({ files: [] }),
    ]).then(([userData, postsData, filesData]) => {
      setProfile(userData)
      setPosts(postsData.posts || [])
      setFiles(filesData.files || [])
      setForm({
        email: userData.email || '',
        bio: userData.bio || '',
        avatar: userData.avatar || '',
        password: '',
        role: userData.role || 'user',
      })
    }).catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }, [profileId])

  const handleUpdate = async (e) => {
    e.preventDefault()
    setError('')
    setMsg('')
    try {
      await api.updateProfile(form)
      setMsg('Profile updated!')
      setEditMode(false)
      const updated = await api.getUser(profileId)
      setProfile(updated)
    } catch (e) {
      setError(e.message)
    }
  }

  const handleFileUpload = async () => {
    if (!uploadFile) return
    try {
      const data = await api.uploadFile(uploadFile)
      setMsg(`File uploaded: ${data.filename}`)
      const filesData = await api.listFiles()
      setFiles(filesData.files || [])
    } catch (e) {
      setError(e.message)
    }
  }

  if (loading) return <div className="loading">Loading profile…</div>
  if (!profile) return <div className="container"><div className="error-msg">{error || 'User not found'}</div></div>

  return (
    <div className="container">
      <div className="main-content">
        {/* Profile header */}
        <div className="card" style={{ padding: 24, marginBottom: 16 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 16 }}>
            <div style={{
              width: 64, height: 64, borderRadius: '50%', background: '#ff4500',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              fontSize: 28, color: '#fff', fontWeight: 700, overflow: 'hidden'
            }}>
              {profile.avatar
                ? <img src={profile.avatar} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                : profile.username[0].toUpperCase()}
            </div>
            <div>
              <h2 style={{ fontSize: 20 }}>u/{profile.username}</h2>
              <div style={{ fontSize: 13, color: '#878a8c' }}>
                Role: <strong>{profile.role}</strong>
              </div>
            </div>
            {isOwn && (
              <button
                className="btn-outline btn-sm"
                style={{ marginLeft: 'auto' }}
                onClick={() => setEditMode(!editMode)}
              >
                {editMode ? 'Cancel' : 'Edit Profile'}
              </button>
            )}
          </div>

          {/* User bio rendered with markdown support */}
          {profile.bio && !editMode && (
            <div
              style={{ fontSize: 14, lineHeight: 1.6 }}
              dangerouslySetInnerHTML={{ __html: profile.bio }}
            />
          )}

          {msg && <div className="success-msg">{msg}</div>}
          {error && <div className="error-msg">{error}</div>}

          {editMode && (
            <form onSubmit={handleUpdate} style={{ marginTop: 16 }}>
              <div style={{ display: 'grid', gap: 12 }}>
                <div>
                  <label style={{ fontSize: 12, fontWeight: 700, display: 'block', marginBottom: 4 }}>EMAIL</label>
                  <input type="email" value={form.email} onChange={e => setForm({ ...form, email: e.target.value })} />
                </div>
                <div>
                  <label style={{ fontSize: 12, fontWeight: 700, display: 'block', marginBottom: 4 }}>
                    BIO <span style={{ fontWeight: 400, color: '#878a8c' }}>(HTML supported)</span>
                  </label>
                  <textarea value={form.bio} onChange={e => setForm({ ...form, bio: e.target.value })} placeholder="Tell us about yourself…" />
                </div>
                <div>
                  <label style={{ fontSize: 12, fontWeight: 700, display: 'block', marginBottom: 4 }}>AVATAR URL</label>
                  <input type="text" value={form.avatar} onChange={e => setForm({ ...form, avatar: e.target.value })} placeholder="https://…" />
                </div>
                <div>
                  <label style={{ fontSize: 12, fontWeight: 700, display: 'block', marginBottom: 4 }}>NEW PASSWORD</label>
                  <input type="password" value={form.password} onChange={e => setForm({ ...form, password: e.target.value })} placeholder="Leave blank to keep current" />
                </div>
                <div>
                  <label style={{ fontSize: 12, fontWeight: 700, display: 'block', marginBottom: 4 }}>ROLE</label>
                  <select value={form.role} onChange={e => setForm({ ...form, role: e.target.value })} style={{ width: 'auto' }}>
                    <option value="user">user</option>
                    <option value="moderator">moderator</option>
                    <option value="admin">admin</option>
                  </select>
                </div>
              </div>
              <button type="submit" className="btn-primary" style={{ marginTop: 16 }}>Save</button>
            </form>
          )}
        </div>

        {/* File manager */}
        {isOwn && (
          <div className="card" style={{ padding: 16, marginBottom: 16 }}>
            <h3 style={{ marginBottom: 12 }}>My Files</h3>
            <div style={{ display: 'flex', gap: 8, marginBottom: 12 }}>
              <input
                type="file"
                onChange={e => setUploadFile(e.target.files[0])}
                style={{ flex: 1 }}
              />
              <button className="btn-primary btn-sm" onClick={handleFileUpload}>Upload</button>
            </div>
            {files.length === 0 && <p style={{ fontSize: 13, color: '#878a8c' }}>No files uploaded yet.</p>}
            {files.map(f => (
              <div key={f.id} style={{ display: 'flex', justifyContent: 'space-between', padding: '6px 0', borderBottom: '1px solid #edeff1', fontSize: 13 }}>
                <span>{f.original_name}</span>
                <a href={f.download_url} target="_blank" rel="noreferrer">Download</a>
              </div>
            ))}
          </div>
        )}

        {/* User's posts */}
        <h3 style={{ marginBottom: 10 }}>Posts by u/{profile.username}</h3>
        {posts.length === 0 && (
          <div className="card" style={{ padding: 24, textAlign: 'center', color: '#878a8c' }}>
            No posts yet.
          </div>
        )}
        {posts.map(p => <PostCard key={p.id} post={p} />)}
      </div>

      <aside className="sidebar">
        <div className="card" style={{ padding: 16 }}>
          <h4 style={{ marginBottom: 8 }}>Account Info</h4>
          <div style={{ fontSize: 13 }}>
            <div style={{ padding: '4px 0', borderBottom: '1px solid #edeff1' }}>
              <span style={{ color: '#878a8c' }}>Username: </span>{profile.username}
            </div>
            <div style={{ padding: '4px 0', borderBottom: '1px solid #edeff1' }}>
              <span style={{ color: '#878a8c' }}>Email: </span>{profile.email}
            </div>
            <div style={{ padding: '4px 0' }}>
              <span style={{ color: '#878a8c' }}>Member since: </span>
              {new Date(profile.created_at).toLocaleDateString()}
            </div>
          </div>
        </div>
      </aside>
    </div>
  )
}
