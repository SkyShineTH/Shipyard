import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      '/api/v1/todos': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/api/v1/register': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/api/v1/login': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
    },
  },
})
