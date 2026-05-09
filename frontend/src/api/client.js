const BASE = '/api'

function headers(extra = {}) {
  return { 'Content-Type': 'application/json', ...extra }
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
  logout:   (next) => request('POST', '/auth/logout', next ? { next } : undefined),
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

  // File upload (multipart) — credentials: 'include' sends the HttpOnly session cookie
  uploadFile: async (file) => {
    const form = new FormData()
    form.append('file', file)
    const res = await fetch(BASE + '/files/upload', {
      method: 'POST',
      credentials: 'include',
      body: form,
    })
    const data = await res.json().catch(() => ({}))
    if (!res.ok) throw new Error(data.error || 'Upload failed')
    return data
  },
}
