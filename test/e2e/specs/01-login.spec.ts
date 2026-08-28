import { test } from '@playwright/test'
import { fillCredentials, signIn } from '../helpers/auth'
import { shoot } from '../helpers/screenshot'
import { required } from '../helpers/env'

test('authelia login form is served', async ({ page }, testInfo) => {
  await page.goto(`https://auth.${required('PLAYWRIGHT_FULL_DOMAIN')}`)
  await fillCredentials(page)
  await shoot(page, testInfo, 'auth')
})

test('sign in through syncloud sso and reach the collabora admin console', async ({ page }, testInfo) => {
  await page.goto('/')
  await shoot(page, testInfo, 'login')
  await signIn(page)
  await shoot(page, testInfo, 'admin')
})
