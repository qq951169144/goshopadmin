import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd())
  
  return {
    plugins: [vue()],
    server: {
      port: 3001,
      proxy: {
        '/api/search': {
          target: env.VITE_SEARCH_PROXY_TARGET || 'http://search-service:8082',
          changeOrigin: true
        },
        '/api': {
          target: env.VITE_API_PROXY_TARGET || 'http://localhost:8081',
          changeOrigin: true
        }
      }
    }
  }
})