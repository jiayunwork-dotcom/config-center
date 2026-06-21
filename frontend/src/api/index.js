import axios from 'axios'

const API_BASE = '/api/v1'

const TOKEN_KEY = 'config_center_token'
const USER_KEY = 'config_center_user'

export const getToken = () => localStorage.getItem(TOKEN_KEY)
export const setToken = (token) => localStorage.setItem(TOKEN_KEY, token)
export const clearToken = () => {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}
export const getUser = () => {
  const raw = localStorage.getItem(USER_KEY)
  return raw ? JSON.parse(raw) : null
}
export const setUser = (user) => localStorage.setItem(USER_KEY, JSON.stringify(user))

const api = axios.create({
  baseURL: API_BASE,
  timeout: 30000,
})

api.interceptors.request.use(
  (config) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

api.interceptors.response.use(
  response => response.data,
  error => {
    console.error('API Error:', error)
    if (error.response && error.response.status === 401) {
      clearToken()
      if (!window.location.pathname.includes('/login')) {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  }
)

export const authApi = {
  login: (username, password) => api.post('/auth/login', { username, password }),
  me: () => api.get('/auth/me'),
}

export const userApi = {
  list: () => api.get('/users'),
  create: (username, password) => api.post('/users', { username, password }),
  delete: (id) => api.delete(`/users/${id}`),
  getRoles: (userId) => api.get(`/users/${userId}/roles`),
}

export const roleApi = {
  grant: (user_id, namespace_id, role) => api.post('/roles/grant', { user_id, namespace_id, role }),
  revoke: (id) => api.delete(`/roles/${id}`),
}

export const auditApi = {
  list: (params) => api.get('/audit-logs', { params }),
}

export const namespaceApi = {
  list: () => api.get('/namespaces'),
  create: (data) => api.post('/namespaces', data),
  update: (id, data) => api.put(`/namespaces/${id}`, data),
  delete: (id) => api.delete(`/namespaces/${id}`),
  get: (id) => api.get(`/namespaces/${id}`)
}

export const groupApi = {
  list: (namespaceId) => api.get('/groups', { params: { namespace_id: namespaceId } }),
  create: (data) => api.post('/groups', data),
  update: (id, data) => api.put(`/groups/${id}`, data),
  delete: (id) => api.delete(`/groups/${id}`),
  get: (id) => api.get(`/groups/${id}`)
}

export const configApi = {
  list: (params) => api.get('/configs', { params }),
  get: (id) => api.get(`/configs/${id}`),
  create: (data) => api.post('/configs', data),
  update: (id, data) => api.put(`/configs/${id}`, data),
  delete: (id) => api.delete(`/configs/${id}`),
  validate: (data) => api.post('/configs/validate', data),
  getMerged: (params) => api.get('/configs/merged', { params }),
  rollback: (id, version) => api.post(`/configs/${id}/rollback`, { version }),
  getVersions: (id, params) => api.get(`/configs/${id}/versions`, { params }),
  compareVersions: (id, version1, version2) => 
    api.get(`/configs/${id}/compare`, { params: { version1, version2 } }),
  batchDelete: (ids) => api.post('/configs/batch-delete', { ids }),
  batchCopy: (sourceIds, targetEnvironment) => 
    api.post('/configs/batch-copy', { source_ids: sourceIds, target_environment: targetEnvironment })
}

export const grayApi = {
  list: (configItemId) => api.get('/gray', { params: { config_item_id: configItemId } }),
  create: (data) => api.post('/gray', data),
  get: (id) => api.get(`/gray/${id}`),
  start: (id) => api.post(`/gray/${id}/start`),
  fullPush: (id) => api.post(`/gray/${id}/full-push`),
  rollback: (id) => api.post(`/gray/${id}/rollback`)
}

export const pushApi = {
  longPoll: (params) => api.get('/push/long-poll', { params, timeout: 35000 }),
  connections: (namespaceId) => api.get('/push/connections', { params: { namespace_id: namespaceId } }),
  stats: () => api.get('/push/stats')
}

export const metricApi = {
  get: (params) => api.get('/metrics', { params }),
  latest: (namespaceId) => api.get('/metrics/latest', { params: { namespace_id: namespaceId } })
}

export const permissionApi = {
  getRoleForNamespace: (me, namespaceId) => {
    if (!me) return 'viewer'
    if (me.is_global_admin) return 'admin'
    let highest = 0
    let role = ''
    const levels = { viewer: 1, editor: 2, admin: 3 }
    for (const r of me.roles || []) {
      if (r.namespace_id === null || r.namespace_id === namespaceId) {
        const lvl = levels[r.role] || 0
        if (lvl > highest) {
          highest = lvl
          role = r.role
        }
      }
    }
    return role
  },
  canEdit: (me, namespaceId) => {
    const role = permissionApi.getRoleForNamespace(me, namespaceId)
    return role === 'editor' || role === 'admin'
  },
  isAdmin: (me) => me && me.is_global_admin
}

export default api
