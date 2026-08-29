import { test, expect } from '@playwright/test'
import { signIn } from '../helpers/auth'
import { shoot } from '../helpers/screenshot'
import { createDocument } from '../helpers/editor'

test('create a document, see it listed, then delete it', async ({ page }, testInfo) => {
  const name = 'files-spec.docx'

  await signIn(page)
  await createDocument(page, 'docx', name)
  await page.getByTestId('editor-back').click()

  await expect(page.getByTestId(`file-row-${name}`)).toBeVisible()
  await shoot(page, testInfo, 'files-list')

  await page.getByTestId(`file-delete-${name}`).click()
  await expect(page.getByTestId(`file-row-${name}`)).toHaveCount(0)
  await shoot(page, testInfo, 'files-after-delete')
})
