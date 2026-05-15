import axios from 'axios'

const api = axios.create({
  baseURL: '',
  timeout: 30000,
  headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
})

api.interceptors.request.use(config => {
  const t = localStorage.getItem('oj_token')
  if (t) config.headers.Authorization = t
  return config
})

export default api
