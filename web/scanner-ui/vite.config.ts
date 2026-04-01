import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

export default defineConfig({
  base: '/scanner/',
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    outDir: '../../internal/scanner/web',
    emptyOutDir: true,
  },
  server: {
    proxy: {
      '/scanner/api': {
        target: 'http://localhost:8101',
        changeOrigin: true,
      },
    },
  },
})
