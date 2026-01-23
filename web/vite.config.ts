import path from 'path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import tailwindcss from '@tailwindcss/vite'
import { tanstackRouter } from '@tanstack/router-plugin/vite'

export default defineConfig({
  plugins: [
    tanstackRouter({
      target: 'react',
      autoCodeSplitting: true,
    }),
    react(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    chunkSizeWarningLimit: 600, // react-dom 固有大小 ~500KB
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            if (id.includes('react-dom') || id.includes('/react/')) {
              return 'vendor-react'
            }
            if (id.includes('@tanstack/react-router')) {
              return 'vendor-router'
            }
            if (id.includes('@tanstack/react-query')) {
              return 'vendor-query'
            }
            if (id.includes('@radix-ui')) {
              return 'vendor-radix'
            }
            if (id.includes('recharts') || id.includes('d3-')) {
              return 'vendor-recharts'
            }
          }
        },
      },
    },
  },
  server: {
    watch: {
      ignored: ['**/routeTree.gen.ts'],
    },
    proxy: {
      '/api/v1': {
        target: 'http://localhost:18080',
        changeOrigin: true,
      },
    },
  },
})
