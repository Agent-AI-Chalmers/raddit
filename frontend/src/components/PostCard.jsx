import React from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import { useAuth } from '../contexts/AuthContext'
import { sanitizeHTML } from '../utils/sanitize'

export default function PostCard({ post, onDelete }) {
  const { user } = useAuth()

  const handleVote = async (type) => {
    try { await api.votePost(post.id, type) } catch (e) { /* silent */ }
  }

  const handleDelete = async () => {
    if (!window.confirm('Delete this post?')) return
    try { await api.deletePost(post.id); onDelete && onDelete(post.id) }
    catch (e) { alert(e.message) }
  }

  const timeAgo = (dateStr) => {
    const diff = Date.now() - new Date(dateStr).getTime()
    const m = Math.floor(diff / 60000)
    if (m < 60) return `${m}m ago`
    if (m < 1440) return `${Math.floor(m / 60)}h ago`
    return `${Math.floor(m / 1440)}d ago`
  }

  return (
    <div className="card" style={{ display: 'flex', marginBottom: 10 }}>
      {/* Vote column */}
      <div style={{
        background: '#f8f9fa', width: 40, display: 'flex', flexDirection: 'column',
        alignItems: 'center', padding: '8px 4px', gap: 4, borderRadius: '4px 0 0 4px'
      }}>
        <button
          onClick={() => handleVote('up')}
          style={{ background: 'none', color: '#878a8c', fontSize: 18, padding: 2 }}
          title="Upvote"
        >▲</button>
        <span style={{ fontSize: 12, fontWeight: 700 }}>{post.upvotes - post.downvotes}</span>
        <button
          onClick={() => handleVote('down')}
          style={{ background: 'none', color: '#878a8c', fontSize: 18, padding: 2 }}
          title="Downvote"
        >▼</button>
      </div>

      {/* Content */}
      <div style={{ padding: '8px 12px', flex: 1 }}>
        <div style={{ fontSize: 12, color: '#878a8c', marginBottom: 4 }}>
          <span style={{ fontWeight: 700, color: '#1c1c1c' }}>r/{post.subreddit}</span>
          {' · Posted by '}
          <Link to={`/user/${post.user_id}`}>u/{post.username}</Link>
          {' · '}
          {timeAgo(post.created_at)}
        </div>

        <Link to={`/posts/${post.id}`} style={{ color: '#1c1c1c', fontSize: 18, fontWeight: 500 }}>
          {post.title}
        </Link>

        {/* Render content preview — supports rich text (sanitized) */}
        <div
          style={{ fontSize: 14, color: '#3c3c3c', marginTop: 6, maxHeight: 80, overflow: 'hidden' }}
          dangerouslySetInnerHTML={{ __html: sanitizeHTML(post.content) }}
        />

        <div style={{ display: 'flex', gap: 12, marginTop: 8, fontSize: 12, color: '#878a8c' }}>
          <Link to={`/posts/${post.id}`} style={{ color: '#878a8c' }}>
            💬 Comments
          </Link>
          {user && (user.id === post.user_id || user.role === 'admin') && (
            <>
              <Link to={`/posts/${post.id}/edit`} style={{ color: '#878a8c' }}>Edit</Link>
              <button
                onClick={handleDelete}
                style={{ background: 'none', color: '#ea0027', fontSize: 12, padding: 0 }}
              >Delete</button>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
