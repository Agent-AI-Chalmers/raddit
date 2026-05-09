import React, { useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import { useAuth } from '../contexts/AuthContext'

export default function Comment({ comment, onDelete, onReply }) {
  const { user } = useAuth()
  const [replying, setReplying] = useState(false)
  const [replyText, setReplyText] = useState('')

  const timeAgo = (dateStr) => {
    const diff = Date.now() - new Date(dateStr).getTime()
    const m = Math.floor(diff / 60000)
    if (m < 60) return `${m}m ago`
    if (m < 1440) return `${Math.floor(m / 60)}h ago`
    return `${Math.floor(m / 1440)}d ago`
  }

  const handleDelete = async () => {
    if (!window.confirm('Delete this comment?')) return
    try { await api.deleteComment(comment.id); onDelete && onDelete(comment.id) }
    catch (e) { alert(e.message) }
  }

  const submitReply = async (e) => {
    e.preventDefault()
    if (!replyText.trim()) return
    try {
      await api.createComment({ content: replyText, post_id: comment.post_id, parent_id: comment.id })
      setReplyText('')
      setReplying(false)
      onReply && onReply()
    } catch (e) { alert(e.message) }
  }

  return (
    <div style={{ borderLeft: '2px solid #edeff1', paddingLeft: 12, marginBottom: 12 }}>
      <div style={{ fontSize: 12, color: '#878a8c', marginBottom: 4 }}>
        <Link to={`/user/${comment.user_id}`} style={{ fontWeight: 700 }}>
          u/{comment.username}
        </Link>
        {' · '}{timeAgo(comment.created_at)}
      </div>

      {/* Render comment content as text to prevent XSS */}
      <div style={{ fontSize: 14, lineHeight: 1.5, whiteSpace: 'pre-wrap' }}>
        {comment.content}
      </div>

      <div style={{ display: 'flex', gap: 10, marginTop: 6, fontSize: 12, color: '#878a8c' }}>
        {user && (
          <button
            onClick={() => setReplying(!replying)}
            style={{ background: 'none', color: '#878a8c', padding: 0, fontSize: 12 }}
          >Reply</button>
        )}
        {user && (user.id === comment.user_id || user.role === 'admin') && (
          <button
            onClick={handleDelete}
            style={{ background: 'none', color: '#ea0027', padding: 0, fontSize: 12 }}
          >Delete</button>
        )}
      </div>

      {replying && (
        <form onSubmit={submitReply} style={{ marginTop: 8 }}>
          <textarea
            value={replyText}
            onChange={e => setReplyText(e.target.value)}
            placeholder="Write a reply…"
            style={{ minHeight: 80 }}
          />
          <div style={{ display: 'flex', gap: 8, marginTop: 6 }}>
            <button type="submit" className="btn-primary btn-sm">Reply</button>
            <button type="button" className="btn-outline btn-sm" onClick={() => setReplying(false)}>Cancel</button>
          </div>
        </form>
      )}
    </div>
  )
}
