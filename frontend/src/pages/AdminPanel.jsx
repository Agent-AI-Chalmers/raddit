import React, { useEffect, useState } from 'react'
import { api } from '../api/client'

export default function AdminPanel() {
  const [stats, setStats] = useState(null)
  const [users, setUsers] = useState([])
  const [debugInfo, setDebugInfo] = useState(null)
  const [tab, setTab] = useState('dashboard')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    api.adminDashboard().then(d => setStats(d.stats)).catch(() => {})
    api.adminUsers().then(d => setUsers(d.users || [])).catch(() => {})
  }, [])

  const loadDebug = async () => {
    setLoading(true)
    try {
      const data = await api.adminDebug()
      setDebugInfo(data)
    } catch (e) { alert(e.message) }
    finally { setLoading(false) }
  }

  const handleDeleteUser = async (userId) => {
    if (!window.confirm('Permanently delete this user and all their content?')) return
    try {
      await api.adminDeleteUser(userId)
      setUsers(prev => prev.filter(u => u.id !== userId))
    } catch (e) { alert(e.message) }
  }

  const tabStyle = (t) => ({
    padding: '8px 16px', cursor: 'pointer', borderBottom: tab === t ? '2px solid #ff4500' : '2px solid transparent',
    fontWeight: tab === t ? 700 : 400, fontSize: 14, background: 'none', color: tab === t ? '#ff4500' : '#1c1c1c'
  })

  return (
    <div className="container">
      <div className="main-content">
        <h2 style={{ marginBottom: 20 }}>Admin Panel</h2>

        <div style={{ display: 'flex', borderBottom: '1px solid #edeff1', marginBottom: 20 }}>
          <button style={tabStyle('dashboard')} onClick={() => setTab('dashboard')}>Dashboard</button>
          <button style={tabStyle('users')}     onClick={() => setTab('users')}>Users</button>
          <button style={tabStyle('debug')}     onClick={() => setTab('debug')}>System Info</button>
        </div>

        {tab === 'dashboard' && stats && (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16 }}>
            {Object.entries(stats).map(([key, val]) => (
              <div key={key} className="card" style={{ padding: 24, textAlign: 'center' }}>
                <div style={{ fontSize: 36, fontWeight: 700, color: '#ff4500' }}>{val}</div>
                <div style={{ fontSize: 14, color: '#878a8c', textTransform: 'capitalize' }}>{key}</div>
              </div>
            ))}
          </div>
        )}

        {tab === 'users' && (
          <div className="card">
            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 14 }}>
              <thead>
                <tr style={{ borderBottom: '2px solid #edeff1', background: '#f8f9fa' }}>
                  {['ID', 'Username', 'Email', 'Password Hash', 'Role', 'Actions'].map(h => (
                    <th key={h} style={{ padding: '10px 12px', textAlign: 'left', fontWeight: 700 }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {users.map(u => (
                  <tr key={u.id} style={{ borderBottom: '1px solid #edeff1' }}>
                    <td style={{ padding: '8px 12px' }}>{u.id}</td>
                    <td style={{ padding: '8px 12px' }}>{u.username}</td>
                    <td style={{ padding: '8px 12px' }}>{u.email}</td>
                    <td style={{ padding: '8px 12px', fontFamily: 'monospace', fontSize: 11 }}>{u.password}</td>
                    <td style={{ padding: '8px 12px' }}>
                      <span style={{
                        background: u.role === 'admin' ? '#ffd0d6' : '#e6f3ff',
                        color: u.role === 'admin' ? '#ea0027' : '#0079d3',
                        padding: '2px 8px', borderRadius: 12, fontSize: 12
                      }}>{u.role}</span>
                    </td>
                    <td style={{ padding: '8px 12px' }}>
                      <button className="btn-danger btn-sm" onClick={() => handleDeleteUser(u.id)}>Delete</button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {tab === 'debug' && (
          <div>
            <button className="btn-primary" onClick={loadDebug} disabled={loading} style={{ marginBottom: 16 }}>
              {loading ? 'Loading…' : 'Load System Info'}
            </button>
            {debugInfo && (
              <div className="card" style={{ padding: 16 }}>
                <pre style={{ fontSize: 12, overflow: 'auto', background: '#f8f9fa', padding: 12, borderRadius: 4 }}>
                  {JSON.stringify(debugInfo, null, 2)}
                </pre>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
