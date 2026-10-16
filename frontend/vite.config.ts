import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  // Pre-bundle the map libraries at startup rather than on first page load.
  optimizeDeps: {
    include: ['maplibre-gl', 'react-map-gl/maplibre'],
  },
  server: {
    host: true,
    port: 5173,
    proxy: {
      // The backend owns all data access; the browser never talks to Martin.
      '/api': process.env.API_PROXY_TARGET ?? 'http://localhost:8080',
    },
  },
});
