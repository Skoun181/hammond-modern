import { fileURLToPath } from 'node:url'
import vue from '@vitejs/plugin-vue'
import { defineConfig, loadEnv } from 'vite'

const aliases = {
  '@src': './src',
  '@router': './src/router',
  '@views': './src/router/views',
  '@layouts': './src/router/layouts',
  '@components': './src/components',
  '@assets': './src/assets',
  '@utils': './src/utils',
  '@state': './src/state',
  '@design': './src/design/index.scss',
  '@': './',
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  return {
    plugins: [vue()],
    resolve: {
      alias: Object.fromEntries(
        Object.entries(aliases).map(([alias, path]) => [
          alias,
          fileURLToPath(new URL(path, import.meta.url)),
        ])
      ),
    },
    server: {
      port: 8080,
      proxy: {
        '/api': env.API_BASE_URL || 'http://localhost:3000',
      },
    },
    test: {
      environment: 'jsdom',
      globals: true,
      maxWorkers: 4,
      include: ['src/**/*.unit.js'],
      setupFiles: ['./tests/unit/setup.js', './tests/unit/matchers.js'],
    },
  }
})
