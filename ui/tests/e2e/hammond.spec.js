import { expect, test } from '@playwright/test'

function field(page, title) {
  return page.locator('.field').filter({ has: page.getByText(title, { exact: true }) })
}

test('sets up, logs in, creates a vehicle and exports a backup', async ({ page }) => {
  const errors = []
  page.on('pageerror', error => errors.push(error.message))
  const email = 'browser-fixture@example.com'
  const password = 'browser-fixture-password-123'

  await page.goto('/')
  await expect(page.getByRole('heading', { name: 'Hammond Modern einrichten' })).toBeVisible()
  await field(page, 'Name').locator('input').fill('Testkonto')
  await page.locator('input[type=email]').fill(email)
  await page.locator('input[type=password]').fill(password)
  await page.getByRole('button', { name: 'Admin-Konto erstellen' }).click()
  await expect(page).toHaveURL(/\/login$/, { timeout: 20_000 })
  await page.locator('input[type=email]').fill(email)
  await page.locator('input[type=password]').fill(password)
  await page.locator('form [type=submit]').click()
  await expect(page.getByRole('heading', { name: 'Dein Dashboard' })).toBeVisible()
  await page.reload()
  await expect(page.getByRole('heading', { name: 'Dein Dashboard' })).toBeVisible()

  await page.route('**/api/vehicle-models?*', route => route.fulfill({ json: { models: ['Corsa'] } }))
  await page.goto('/vehicles/create')
  const form = page.locator('form')
  await form.locator('input[type=text]').nth(0).fill('Testfahrzeug')
  await form.locator('input[type=text]').nth(1).fill('TEST-123')
  await form.locator('select').nth(0).selectOption({ index: 1 })
  await form.locator('select').nth(1).selectOption({ index: 1 })
  await form.getByPlaceholder('Marke suchen oder eingeben').fill('Opel')
  await form.getByPlaceholder('Modell auswählen oder eingeben').fill('Corsa')
  const created = page.waitForResponse(response =>
    new URL(response.url()).pathname === '/api/vehicles' && response.request().method() === 'POST')
  await form.locator('[type=submit]').click()
  const result = await created
  expect(result.status()).toBe(201)
  const vehicle = await result.json()
  await page.goto(`/vehicles/${vehicle.id}`)
  await expect(page.getByText('Testfahrzeug', { exact: false }).first()).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Statistiken' })).toBeVisible()
  await expect(page.locator('canvas')).toBeVisible()

  await page.goto('/admin/backup')
  const [download] = await Promise.all([
    page.waitForEvent('download'),
    page.getByRole('button', { name: /Backup herunterladen/i }).click(),
  ])
  expect(download.suggestedFilename()).toMatch(/\.zip$/)
  expect(errors).toEqual([])
})
