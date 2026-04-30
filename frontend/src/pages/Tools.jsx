import React, { useState } from 'react'
import { api } from '../api/client'

export default function Tools() {
  const [host, setHost] = useState('')
  const [pingResult, setPingResult] = useState('')
  const [nslookupResult, setNslookupResult] = useState('')
  const [previewUrl, setPreviewUrl] = useState('')
  const [previewResult, setPreviewResult] = useState(null)
  const [loading, setLoading] = useState({})

  const run = async (key, fn) => {
    setLoading(p => ({ ...p, [key]: true }))
    try { return await fn() }
    finally { setLoading(p => ({ ...p, [key]: false })) }
  }

  const handlePing = async () => {
    const data = await run('ping', () => api.ping(host))
    setPingResult(data.output || 'No output')
  }

  const handleNslookup = async () => {
    const data = await run('nslookup', () => api.nslookup(host))
    setNslookupResult(data.result || 'No output')
  }

  const handlePreview = async () => {
    const data = await run('preview', () => api.previewURL(previewUrl))
    setPreviewResult(data)
  }

  return (
    <div className="container">
      <div className="main-content">
        <h2 style={{ marginBottom: 20 }}>Network Tools</h2>
        <p style={{ marginBottom: 20, fontSize: 14, color: '#878a8c' }}>
          Utility tools for testing network connectivity and diagnosing issues.
        </p>

        {/* Ping Tool */}
        <div className="card" style={{ padding: 20, marginBottom: 16 }}>
          <h3 style={{ marginBottom: 12 }}>Ping / NSLookup</h3>
          <div style={{ display: 'flex', gap: 8, marginBottom: 12 }}>
            <input
              type="text"
              value={host}
              onChange={e => setHost(e.target.value)}
              placeholder="hostname or IP address"
            />
            <button className="btn-primary btn-sm" onClick={handlePing} disabled={loading.ping}>
              {loading.ping ? '…' : 'Ping'}
            </button>
            <button className="btn-outline btn-sm" onClick={handleNslookup} disabled={loading.nslookup}>
              {loading.nslookup ? '…' : 'NSLookup'}
            </button>
          </div>
          {pingResult && (
            <pre style={{ background: '#1c1c1c', color: '#00ff00', padding: 12, borderRadius: 4, fontSize: 12, overflow: 'auto', maxHeight: 200 }}>
              {pingResult}
            </pre>
          )}
          {nslookupResult && (
            <pre style={{ background: '#1c1c1c', color: '#00ff00', padding: 12, borderRadius: 4, fontSize: 12, overflow: 'auto', maxHeight: 200, marginTop: 8 }}>
              {nslookupResult}
            </pre>
          )}
        </div>

        {/* URL Preview Tool */}
        <div className="card" style={{ padding: 20 }}>
          <h3 style={{ marginBottom: 12 }}>URL Content Preview</h3>
          <div style={{ display: 'flex', gap: 8, marginBottom: 12 }}>
            <input
              type="text"
              value={previewUrl}
              onChange={e => setPreviewUrl(e.target.value)}
              placeholder="https://example.com"
            />
            <button className="btn-primary btn-sm" onClick={handlePreview} disabled={loading.preview}>
              {loading.preview ? '…' : 'Fetch'}
            </button>
          </div>
          {previewResult && (
            <div>
              <div style={{ fontSize: 12, color: '#878a8c', marginBottom: 8 }}>
                Status: {previewResult.status_code} · {previewResult.content_type}
              </div>
              <pre style={{ background: '#f8f9fa', padding: 12, borderRadius: 4, fontSize: 12, overflow: 'auto', maxHeight: 300 }}>
                {previewResult.preview}
              </pre>
            </div>
          )}
        </div>
      </div>

      <aside className="sidebar">
        <div className="card" style={{ padding: 16 }}>
          <h4 style={{ marginBottom: 8 }}>About Tools</h4>
          <p style={{ fontSize: 13, color: '#3c3c3c' }}>
            These network utilities help diagnose connectivity issues and preview external content.
          </p>
        </div>
      </aside>
    </div>
  )
}
