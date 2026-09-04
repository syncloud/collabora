import { test, expect } from '@playwright/test'
import { signIn } from '../helpers/auth'
import { shoot } from '../helpers/screenshot'
import { fetchDocument, waitEditorLoaded } from '../helpers/editor'

const upgradeDocument = 'upgrade.txt'
const upgradePhrase = 'SurvivesTheUpgrade'

test('the seeded document survived the refresh and still opens', async ({ page }, testInfo) => {
  await signIn(page)
  await expect(page.getByTestId(`file-row-${upgradeDocument}`)).toBeVisible()
  await shoot(page, testInfo, 'post-upgrade-list')

  const saved = await fetchDocument(page, upgradeDocument)
  expect(saved.toString('utf-8')).toContain(upgradePhrase)

  await page.getByTestId(`file-open-${upgradeDocument}`).click()
  await waitEditorLoaded(page)
  await shoot(page, testInfo, 'post-upgrade-editor')
  await expect(page.getByTestId('editor-error')).toHaveCount(0)
})
