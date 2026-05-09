import React, { useEffect, useState } from 'react'
import { useSearchParams, Link } from 'react-router-dom'
import { api } from '../api/client'
import PostCard from '../components/PostCard'

const SUBREDDITS = ['general', 'technology', 'gaming', 'science', 'news', 'funny']

export default function Home() {
  const [params] = useSearchParams()
  const [posts, setPosts] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [sort, setSort] = useState('created_at')
  const [order, setOrder] = useState('DESC')
  const [subreddit, setSubreddit] = useState('')

  const searchQuery = params.get('q') || ''

  useEffect(() => {
    loadPosts()
  }, [searchQuery, sort, order, subreddit])

  const loadPosts = async () => {
    setLoading(true)
    setError('')
    try {
      let data
      if (searchQuery) {
        data = await api.searchPosts(searchQuery)
        setPosts(data.posts || [])
      } else {
        const qStr = new URLSearchParams({
          ...(subreddit && { subreddit }),
          sort,
          order,
        }).toString()
        data = await api.getPosts(`?${qStr}`)
        setPosts(data.posts || [])
      }
    } catch (e) {
      setError(e.message)
    } finally {
      setLoading(false)
    }
  }

  const removePost = (id) => setPosts(prev => prev.filter(p => p.id !== id))

  return (
    <div className="container">
      <div className="main-content">
        {/* Search results header — displays the original query string */}
        {searchQuery && (
          <div style={{ marginBottom: 16 }}>
            <h2 style={{ fontSize: 18 }}>
              Search results for:{' '}
              <span>{searchQuery}</span>
            </h2>
            <p style={{ fontSize: 13, color: '#878a8c' }}>{posts.length} result(s) found</p>
          </div>
        )}

        {!searchQuery && (
          <div className="card" style={{ padding: '8px 12px', marginBottom: 10, display: 'flex', gap: 8, alignItems: 'center', flexWrap: 'wrap' }}>
            <span style={{ fontSize: 13, fontWeight: 700 }}>Sort:</span>
            <select value={sort} onChange={e => setSort(e.target.value)} style={{ width: 'auto' }}>
              <option value="p.created_at">New</option>
              <option value="p.upvotes">Top</option>
              <option value="p.downvotes">Controversial</option>
              <option value="p.id">Rising</option>
            </select>
            <select value={order} onChange={e => setOrder(e.target.value)} style={{ width: 'auto' }}>
              <option value="DESC">Descending</option>
              <option value="ASC">Ascending</option>
            </select>
          </div>
        )}

        {error && <div className="error-msg">{error}</div>}
        {loading && <div className="loading">Loading posts…</div>}
        {!loading && posts.length === 0 && (
          <div className="card" style={{ padding: 40, textAlign: 'center', color: '#878a8c' }}>
            No posts yet. <Link to="/submit">Be the first to post!</Link>
          </div>
        )}
        {!loading && posts.map(post => (
          <PostCard key={post.id} post={post} onDelete={removePost} />
        ))}
      </div>

      <aside className="sidebar">
        <div className="card" style={{ padding: 16, marginBottom: 16 }}>
          <h3 style={{ fontWeight: 700, marginBottom: 12 }}>Communities</h3>
          {SUBREDDITS.map(s => (
            <div
              key={s}
              onClick={() => setSubreddit(subreddit === s ? '' : s)}
              style={{
                padding: '6px 0',
                cursor: 'pointer',
                fontWeight: subreddit === s ? 700 : 400,
                color: subreddit === s ? '#ff4500' : '#1c1c1c',
                borderBottom: '1px solid #edeff1',
              }}
            >
              r/{s}
            </div>
          ))}
        </div>

        <div className="card" style={{ padding: 16 }}>
          <h3 style={{ fontWeight: 700, marginBottom: 8 }}>Home</h3>
          <p style={{ fontSize: 14, color: '#3c3c3c', marginBottom: 12 }}>
            Your personal Raddit front page. Come here to check in with your favourite communities.
          </p>
          <Link to="/submit">
            <button className="btn-primary" style={{ width: '100%', marginBottom: 8 }}>Create Post</button>
          </Link>
        </div>
      </aside>
    </div>
  )
}
