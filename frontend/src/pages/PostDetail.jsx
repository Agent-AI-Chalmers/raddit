import React, { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { api } from '../api/client'
import { useAuth } from '../contexts/AuthContext'
import Comment from '../components/Comment'

export default function PostDetail() {
  const { id } = useParams()
  const { user } = useAuth()
  const [post, setPost] = useState(null)
  const [comments, setComments] = useState([])
  const [newComment, setNewComment] = useState('')
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  useEffect(() => {
    loadPost()
    loadComments()
  }, [id])

  const loadPost = async () => {
    try {
      const data = await api.getPost(id)
      setPost(data)
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  const loadComments = async () => {
    try {
      const data = await api.getComments(id)
      setComments(data.comments || [])
    } catch (e) { /* silent */ }
  }

  const handleVote = async (type) => {
    try {
      const data = await api.votePost(id, type)
      setPost(prev => ({ ...prev, upvotes: data.upvotes, downvotes: data.downvotes }))
    } catch (e) { alert(e.message) }
  }

  const handleComment = async (e) => {
    e.preventDefault()
    if (!newComment.trim()) return
    setSubmitting(true)
    try {
      await api.createComment({ content: newComment, post_id: parseInt(id) })
      setNewComment('')
      loadComments()
    } catch (e) {
      alert(e.message)
    } finally {
      setSubmitting(false)
    }
  }

  const removeComment = (cid) => setComments(prev => prev.filter(c => c.id !== cid))

  if (loading) return <div className="loading">Loading…</div>
  if (error) return <div className="container"><div className="error-msg">{error}</div></div>
  if (!post) return null

  return (
    <div className="container">
      <div className="main-content">
        <div className="card" style={{ display: 'flex', marginBottom: 10 }}>
          {/* Vote column */}
          <div style={{
            background: '#f8f9fa', width: 40, display: 'flex', flexDirection: 'column',
            alignItems: 'center', padding: '8px 4px', gap: 4, borderRadius: '4px 0 0 4px'
          }}>
            <button
              onClick={() => handleVote('up')}
              style={{ background: 'none', color: '#878a8c', fontSize: 18, padding: 2 }}
            >▲</button>
            <span style={{ fontSize: 12, fontWeight: 700 }}>{post.upvotes - post.downvotes}</span>
            <button
              onClick={() => handleVote('down')}
              style={{ background: 'none', color: '#878a8c', fontSize: 18, padding: 2 }}
            >▼</button>
          </div>

          <div style={{ padding: 16, flex: 1 }}>
            <div style={{ fontSize: 12, color: '#878a8c', marginBottom: 8 }}>
              <strong>r/{post.subreddit}</strong>
              {' · Posted by '}
              <Link to={`/user/${post.user_id}`}>u/{post.username}</Link>
            </div>
            <h1 style={{ fontSize: 20, marginBottom: 12 }}>{post.title}</h1>

            {/* Full post content rendered with rich text support */}
            <div
              style={{ fontSize: 14, lineHeight: 1.8 }}
              dangerouslySetInnerHTML={{ __html: post.content }}
            />

            {user && (user.id === post.user_id || user.role === 'admin') && (
              <div style={{ marginTop: 12, display: 'flex', gap: 8 }}>
                <Link to={`/posts/${post.id}/edit`}>
                  <button className="btn-outline btn-sm">Edit</button>
                </Link>
              </div>
            )}
          </div>
        </div>

        {/* Comment section */}
        <div className="card" style={{ padding: 16 }}>
          <h3 style={{ marginBottom: 16 }}>Comments ({comments.length})</h3>

          {user ? (
            <form onSubmit={handleComment} style={{ marginBottom: 20 }}>
              <textarea
                value={newComment}
                onChange={e => setNewComment(e.target.value)}
                placeholder="Share your thoughts…"
                style={{ minHeight: 100 }}
              />
              <button
                type="submit"
                className="btn-primary btn-sm"
                style={{ marginTop: 8 }}
                disabled={submitting}
              >
                {submitting ? 'Posting…' : 'Comment'}
              </button>
            </form>
          ) : (
            <p style={{ marginBottom: 16, fontSize: 14, color: '#878a8c' }}>
              <Link to="/login">Log in</Link> to leave a comment.
            </p>
          )}

          {comments.length === 0 && (
            <p style={{ color: '#878a8c', fontSize: 14 }}>No comments yet. Be the first!</p>
          )}

          {comments.map(c => (
            <Comment
              key={c.id}
              comment={c}
              onDelete={removeComment}
              onReply={loadComments}
            />
          ))}
        </div>
      </div>

      <aside className="sidebar">
        <div className="card" style={{ padding: 16 }}>
          <h4 style={{ marginBottom: 8 }}>About r/{post.subreddit}</h4>
          <p style={{ fontSize: 13, color: '#878a8c' }}>
            Community posts and discussions.
          </p>
        </div>
      </aside>
    </div>
  )
}
