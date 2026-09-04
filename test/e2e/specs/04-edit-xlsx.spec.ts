import { test, expect } from '@playwright/test'
import { signIn } from '../helpers/auth'
import { shoot } from '../helpers/screenshot'
import { createDocument, waitEditorLoaded } from '../helpers/editor'

test('create and open a spreadsheet in the editor', async ({ page }, testInfo) => {
  await signIn(page)
  await createDocument(page, 'xlsx', 'edit-spec.xlsx')
  await waitEditorLoaded(page)
  await shoot(page, testInfo, 'editor-xlsx')
  await expect(page.getByTestId('editor-error')).toHaveCount(0)
})
