import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  // Load env file based on `mode` in the current working directory.
  const env = loadEnv(mode, process.cwd(), '')

  return {
    plugins: [react(), tailwindcss()],
    server: {
      proxy: env.VITE_API_BASE_URL ? undefined : {
        '/api': {
          target: 'http://192.168.1.21:3000',
          changeOrigin: true,
        }
      }
    },
    resolve: {
      alias: {
        "@": path.resolve(__dirname, 'src')
      }
    }
  }
})
