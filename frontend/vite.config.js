import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 3000,
    proxy: {
      // 代理所有 API 请求到 Go 后端
      '^/(login|register|sendcode|problem-list|problem-detail|user-basic|rank-list|submit-list|user|admin|swagger)': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})
