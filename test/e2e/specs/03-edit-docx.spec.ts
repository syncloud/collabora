import { test, expect } from '@playwright/test'
import { signIn } from '../helpers/auth'
import { shoot } from '../helpers/screenshot'
import { createDocument, waitEditorLoaded } from '../helpers/editor'

test('create and open a document in the editor', async ({ page }, testInfo) => {
  await signIn(page)
  await createDocument(page, 'docx', 'edit-spec.docx')
  await waitEditorLoaded(page)
  await shoot(page, testInfo, 'editor-docx')
  await expect(page.getByTestId('editor-error')).toHaveCount(0)
})
