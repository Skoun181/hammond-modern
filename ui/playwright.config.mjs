import { defineConfig } from '@playwright/test'

const baseURL = 'http://127.0.0.1:3000'

export default defineConfig({
  testDir: './tests/e2e',
  timeout: 60_000,
  use: {
    baseURL,
    launchOptions: { executablePath: process.env.HAMMOND_TEST_CHROMIUM },
  },
  webServer: {
    command: 'node ./tests/e2e/server.mjs',
    url: `${baseURL}/api/system/status`,
  },
})
