import { spawn } from 'node:child_process'
import { randomBytes } from 'node:crypto'
import { mkdirSync, mkdtempSync, rmSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

if (!process.env.HAMMOND_TEST_BINARY) {
  throw new Error('Build the Go server and set HAMMOND_TEST_BINARY to the binary path.')
}

const directory = mkdtempSync(join(tmpdir(), 'hammond-e2e-'))
const config = join(directory, 'config')
const assets = join(directory, 'assets')
mkdirSync(config)
mkdirSync(assets)

const child = spawn(resolve(process.env.HAMMOND_TEST_BINARY), [], {
  cwd: fileURLToPath(new URL('../../', import.meta.url)),
  env: {
    ...process.env,
    CONFIG: config,
    DATA: assets,
    JWT_SECRET: randomBytes(32).toString('hex'),
    GIN_MODE: 'release',
    SESSION_COOKIE_SECURE: '0',
  },
  stdio: 'inherit',
})

for (const signal of ['SIGINT', 'SIGTERM']) {
  process.on(signal, () => child.kill(signal))
}
child.on('error', error => {
  rmSync(directory, { recursive: true, force: true })
  console.error(error)
  process.exitCode = 1
})
child.on('exit', code => {
  rmSync(directory, { recursive: true, force: true })
  process.exitCode = code || 0
})
