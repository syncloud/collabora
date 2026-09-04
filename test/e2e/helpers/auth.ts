import { Page, expect } from '@playwright/test'
import { required } from './env'

const deviceUser = required('PLAYWRIGHT_DEVICE_USER')
const devicePassword = required('PLAYWRIGHT_DEVICE_PASSWORD')

export async function fillCredentials(page: Page) {
  await page.locator('#username-textfield').waitFor({ state: 'visible', timeout: 60_000 })
  await page.locator('#username-textfield').fill(deviceUser)
  await page.locator('#password-textfield').fill(devicePassword)
}

export async function signIn(page: Page) {
  await page.goto('/')

  const files = page.getByTestId('files-view')
  const username = page.locator('#username-textfield')
  const consent = page.locator('#accept-button')

  await expect(files.or(username).or(consent).first()).toBeVisible({ timeout: 60_000 })

  if (await username.isVisible()) {
    await fillCredentials(page)
    await page.locator('#sign-in-button').click()
    await expect(files.or(consent).first()).toBeVisible({ timeout: 60_000 })
  }

  if (await consent.isVisible()) {
    await consent.click()
  }

  await expect(files).toBeVisible({ timeout: 60_000 })
}
