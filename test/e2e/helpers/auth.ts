import { Page, expect } from '@playwright/test'
import { required } from './env'

const deviceUser = required('PLAYWRIGHT_DEVICE_USER')
const devicePassword = required('PLAYWRIGHT_DEVICE_PASSWORD')

export async function fillCredentials(page: Page) {
  await page.locator('#username-textfield').fill(deviceUser)
  await page.locator('#password-textfield').fill(devicePassword)
}

export async function signIn(page: Page) {
  await fillCredentials(page)
  await page.locator('#sign-in-button').click()
  await expect(page.locator('#active_users_count')).toBeVisible()
}
