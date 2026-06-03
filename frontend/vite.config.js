import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 3000,
    proxy: {
      // 代理所有 API 请求到 Go 后端
      '^/(login|register|sendcode|problem-list|problem-detail|problem-solution|problem-comments|problem-code-shares|code-share-detail|user-basic|rank-list|submit-list|avatar-url|check-in-leaderboard|contest-list|contest-detail|contest-rank|user|admin|swagger)': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      // WebSocket 代理
      '/ws': {
        target: 'http://localhost:8080',
        ws: true,
        changeOrigin: true
      }
    }
  }
})
