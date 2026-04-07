import { defineConfig } from 'vite';
import { resolve } from 'path';

export default defineConfig({
  root: 'src',
  publicDir: 'public',
  server: {
    port: 5173,
    host: '0.0.0.0',
    cors: true,
    origin: 'http://localhost:5173',
  },
  build: {
    outDir: '../dist',
    emptyOutDir: true,
    manifest: true,
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'src/css/main.css'),
        app: resolve(__dirname, 'src/js/main.js'),
      },
      output: {
        entryFileNames: 'assets/js/[name].[hash].js',
        chunkFileNames: 'assets/js/[name].[hash].js',
        assetFileNames: (assetInfo) => {
          if (assetInfo.names?.[0]?.match(/\.(jpg|jpeg|png|gif|svg|webp|avif)$/)) {
            return 'assets/img/[name][extname]';
          }
          return 'assets/css/[name].[hash][extname]';
        },
      },
    },
  },
});
