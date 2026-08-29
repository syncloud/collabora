import { test } from '@playwright/test'
import { fillCredentials, signIn } from '../helpers/auth'
import { shoot } from '../helpers/screenshot'

test('syncloud sso login lands on the document list', async ({ page }, testInfo) => {
  await page.goto('/')
  await fillCredentials(page)
  await shoot(page, testInfo, 'login')
  await page.locator('#sign-in-button').click()
  await shoot(page, testInfo, 'files')
})

test('the signed in user is shown', async ({ page }, testInfo) => {
  await signIn(page)
  await page.getByTestId('current-user').waitFor()
  await shoot(page, testInfo, 'session')
})
