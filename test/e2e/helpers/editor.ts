import { Page, expect } from '@playwright/test'
import { ssh } from './ssh'
import { execFileSync } from 'node:child_process'
import { mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const frameSelector = '[data-testid="editor-frame"]'

export function editorFrame(page: Page) {
  return page.frameLocator(frameSelector)
}

export async function waitEditorLoaded(page: Page) {
  await expect(page.getByTestId('editor-error')).toHaveCount(0)
  await page.locator(frameSelector).waitFor({ state: 'visible', timeout: 60_000 })
  const canvas = editorFrame(page).locator('#document-container').or(editorFrame(page).locator('canvas').first())
  await expect(canvas.first()).toBeVisible({ timeout: 120_000 })
  await page.waitForTimeout(5000)
  await dismissOverlays(page)
}

async function dismissOverlays(page: Page) {
  const welcome = editorFrame(page).locator('.iframe-welcome-wrap')
  for (let attempt = 0; attempt < 5; attempt++) {
    if ((await welcome.count()) === 0 || !(await welcome.first().isVisible())) {
      return
    }
    await page.keyboard.press('Escape')
    await page.waitForTimeout(1000)
  }
}

export async function focusDocument(page: Page) {
  const frame = editorFrame(page)
  const container = frame.locator('#document-container')
  const target = (await container.count()) > 0 ? container : frame.locator('canvas').first()
  await target.click()
  await page.waitForTimeout(2000)
}

export async function createDocument(page: Page, kind: string, name: string) {
  await page.getByTestId('new-document-button').click()
  await page.getByTestId(`new-${kind}-option`).click()
  await page.getByTestId('new-document-name').fill(name)
  await page.getByTestId('new-document-create').click()
  await expect(page.getByTestId('editor-view')).toBeVisible({ timeout: 60_000 })
}

export async function fetchDocument(page: Page, name: string): Promise<Buffer> {
  const id = fileId(name)
  const response = await page.request.get(`/api/files/${id}/contents`)
  expect(response.ok(), `download ${name} failed: ${response.status()}`).toBe(true)
  return Buffer.from(await response.body())
}

export async function waitForContent(
  page: Page,
  name: string,
  entry: string,
  phrase: string,
  timeoutMs = 180_000
) {
  const deadline = Date.now() + timeoutMs
  let last = 'never fetched'
  while (Date.now() < deadline) {
    await page.waitForTimeout(3000)
    const bytes = await fetchDocument(page, name)
    try {
      if (readZipEntry(bytes, entry).includes(phrase)) return
      last = `${bytes.length} bytes, no match`
    } catch (error) {
      last = `unzip failed: ${(error as Error).message}`
    }
  }
  throw new Error(
    `${name} never contained "${phrase}" within ${timeoutMs}ms (${last})\n` +
      `--- backend ---\n${backendLog()}\n--- coolwsd ---\n${serverLog()}`
  )
}

function backendLog(): string {
  return ssh('journalctl -u snap.collabora.backend -n 80 --no-pager', { throw: false })
}

function serverLog(): string {
  return ssh('journalctl -u snap.collabora.server -n 80 --no-pager', { throw: false })
}

export function readZipEntry(bytes: Buffer, entry: string): string {
  const dir = mkdtempSync(join(tmpdir(), 'cool-zip-'))
  const file = join(dir, 'document')
  writeFileSync(file, bytes)
  return execFileSync('unzip', ['-p', file, entry]).toString('utf-8')
}

export function fileId(name: string): string {
  return Buffer.from(name, 'utf-8').toString('base64url')
}
