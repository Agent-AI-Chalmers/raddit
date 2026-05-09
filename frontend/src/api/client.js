const BASE = '/api'

function getToken() {
  return localStorage.getItem('token')
}

function headers(extra = {}) {
  const h = { 'Content-Type': 'application/json', ...extra }
  // Also send Authorization header for non-browser clients; cookie handles browser sessions
  const token = getToken()
  if (token) h['Authorization'] = `Bearer ${token}`
  return h
}

async function request(method, path, body) {
  const opts = { method, headers: headers(), credentials: 'include' }
  if (body !== undefined) opts.body = JSON.stringify(body)
  const res = await fetch(BASE + path, opts)
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || 'Request failed')
  return data
}

export const api = {
  // Auth
  login:    (body) => request('POST', '/auth/login', body),
  register: (body) => request('POST', '/auth/register', body),
  logout:   (next) => request('GET', `/auth/logout${next ? '?next=' + next : ''}`),
  me:       ()     => request('GET', '/auth/me'),

  // Posts
  getPosts:    (params = '') => request('GET', `/posts${params}`),
  searchPosts: (q)           => request('GET', `/posts/search?q=${encodeURIComponent(q)}`),
  getPost:     (id)          => request('GET', `/posts/${id}`),
  createPost:  (body)        => request('POST', '/posts', body),
  updatePost:  (id, body)    => request('PUT', `/posts/${id}`, body),
  deletePost:  (id)          => request('DELETE', `/posts/${id}`),
  votePost:    (id, type)    => request('POST', `/posts/${id}/vote`, { vote_type: type }),
  previewURL:  (url)         => request('GET', `/posts/preview?url=${encodeURIComponent(url)}`),

  // Comments
  getComments:   (postId) => request('GET', `/posts/${postId}/comments`),
  createComment: (body)   => request('POST', '/comments', body),
  deleteComment: (id)     => request('DELETE', `/comments/${id}`),

  // Users
  getUser:      (id)   => request('GET', `/users/${id}`),
  getUserPosts: (id)   => request('GET', `/users/${id}/posts`),
  updateProfile:(body) => request('PUT', '/users/profile', body),

  // Files
  listFiles: () => request('GET', '/files/list'),

  // Tools
  ping:     (host) => request('GET', `/tools/ping?host=${encodeURIComponent(host)}`),
  nslookup: (host) => request('GET', `/tools/nslookup?host=${encodeURIComponent(host)}`),

  // Admin
  adminDashboard: () => request('GET', '/admin/dashboard'),
  adminDebug:     () => request('GET', '/admin/debug'),
  adminUsers:     () => request('GET', '/admin/users'),
  adminDeleteUser:(id) => request('DELETE', `/admin/users/${id}`),

  // File upload (multipart) — credentials: 'include' sends the session cookie
  uploadFile: async (file) => {
    // Client-side extension validation (defense-in-depth; server enforces too)
    const allowedExts = ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.pdf', '.txt', '.csv',
      '.mp3', '.mp4', '.webm', '.ogg', '.wav', '.zip', '.gz', '.tar']
    const ext = file.name.slice(file.name.lastIndexOf('.')).toLowerCase()
    if (!allowedExts.includes(ext)) {
      throw new Error(`File type "${ext}" is not allowed. Allowed: ${allowedExts.join(', ')}`)
    }

    const form = new FormData()
    form.append('file', file)
    const res = await fetch(BASE + '/files/upload', {
      method: 'POST',
      credentials: 'include',
      headers: { Authorization: `Bearer ${getToken()}` },
      body: form,
    })
    const data = await res.json().catch(() => ({}))
    if (!res.ok) throw new Error(data.error || 'Upload failed')
    return data
  },
}
