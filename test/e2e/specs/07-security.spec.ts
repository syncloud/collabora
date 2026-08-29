import { test, expect, request } from '@playwright/test'
import { required } from '../helpers/env'
import { ssh } from '../helpers/ssh'
import { fileId } from '../helpers/editor'

const appDomain = required('PLAYWRIGHT_APP_DOMAIN')

test('the api rejects requests without a session', async () => {
  const context = await request.newContext({ ignoreHTTPSErrors: true, maxRedirects: 0 })
  for (const path of ['/api/session', '/api/files', '/api/editor?id=' + fileId('x.docx')]) {
    const response = await context.get(`https://${appDomain}${path}`)
    expect(response.status(), `${path} should be 401`).toBe(401)
  }
})

test('the admin console is behind syncloud sso', async () => {
  const context = await request.newContext({ ignoreHTTPSErrors: true, maxRedirects: 0 })
  const response = await context.get(`https://${appDomain}/browser/dist/admin/admin.html`)
  expect(response.status()).toBe(302)
  expect(response.headers()['location']).toContain('auth.')
})

test('the nextcloud integration endpoints stay open', async () => {
  const context = await request.newContext({ ignoreHTTPSErrors: true })
  for (const path of ['/hosting/discovery', '/hosting/capabilities']) {
    const response = await context.get(`https://${appDomain}${path}`)
    expect(response.status(), `${path} should be 200`).toBe(200)
  }
})

test('the wopi host is not reachable through the public proxy', async () => {
  const context = await request.newContext({ ignoreHTTPSErrors: true, maxRedirects: 0 })
  const response = await context.get(`https://${appDomain}/wopi/files/${fileId('x.docx')}`)
  expect(response.status()).toBe(404)
  expect(await response.text()).not.toContain('BaseFileName')
})

test('coolwsd and the wopi host only listen on loopback ipv4', async () => {
  const listening = ssh("sh -c 'ss -lntp'", { throw: false })
  for (const port of ['9980', '9981']) {
    const lines = listening.split('\n').filter((line) => line.includes(`:${port}`))
    expect(lines.length, `nothing listening on ${port} in:\n${listening}`).toBeGreaterThan(0)
    for (const line of lines) {
      expect(line).toContain(`127.0.0.1:${port}`)
    }
  }
})

test('the wopi host itself refuses a forged access token', async () => {
  const id = fileId('x.docx')
  const status = ssh(
    `curl -s -o /dev/null -w '%{http_code}' 'http://127.0.0.1:9981/wopi/files/${id}?access_token=Zm9yZ2Vk.Zm9yZ2Vk'`
  ).trim()
  expect(status).toBe('401')
})
