import { test, expect } from '@playwright/test'
import { signIn } from '../helpers/auth'
import { shoot } from '../helpers/screenshot'
import {
  createDocument,
  fetchDocument,
  focusDocument,
  readZipEntry,
  waitEditorLoaded,
  waitForContent,
} from '../helpers/editor'

test('type into a document, it is saved and survives a reopen', async ({ page }, testInfo) => {
  test.skip(testInfo.project.name !== 'desktop', 'typing is a desktop-only flow')

  const name = 'roundtrip.docx'
  const phrase = 'HelloCollaboraRoundtrip'

  await signIn(page)
  await createDocument(page, 'docx', name)
  await waitEditorLoaded(page)

  await focusDocument(page)
  await shoot(page, testInfo, 'roundtrip-focused')
  await page.keyboard.type(phrase, { delay: 40 })
  await shoot(page, testInfo, 'roundtrip-typed')

  await page.getByTestId('editor-back').click()
  await waitForContent(page, name, 'word/document.xml', phrase)

  const saved = await fetchDocument(page, name)
  expect(readZipEntry(saved, 'word/document.xml')).toContain(phrase)

  await page.getByTestId(`file-open-${name}`).click()
  await waitEditorLoaded(page)
  await shoot(page, testInfo, 'roundtrip-reopened')
  await expect(page.getByTestId('editor-error')).toHaveCount(0)
})
