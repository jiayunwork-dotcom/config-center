import axios from 'axios'

const API_BASE = '/api/v1'

const api = axios.create({
  baseURL: API_BASE,
  timeout: 30000,
  headers: {
    'X-Operator': 'admin'
  }
})

api.interceptors.response.use(
  response => response.data,
  error => {
    console.error('API Error:', error)
    return Promise.reject(error)
  }
)

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
    api.get(`/configs/${id}/compare`, { params: { version1, version2 } })
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

export default api
