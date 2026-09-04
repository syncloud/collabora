import { defineConfig, devices } from '@playwright/test'
import { required } from './helpers/env'

const fullDomain = required('PLAYWRIGHT_FULL_DOMAIN')
const appDomain = required('PLAYWRIGHT_APP_DOMAIN')
const artifactDir = required('PLAYWRIGHT_ARTIFACT_DIR')
const deviceIp = required('PLAYWRIGHT_DEVICE_IP')

export default defineConfig({
  testDir: './specs',
  fullyParallel: false,
  workers: 1,
  retries: process.env.CI ? 1 : 0,
  maxFailures: 0,
  reporter: [['list'], ['html', { outputFolder: `${artifactDir}/playwright/report`, open: 'never' }]],
  outputDir: `${artifactDir}/playwright/test-results`,
  globalSetup: './globalSetup.ts',
  globalTeardown: './globalTeardown.ts',
  timeout: 300_000,
  expect: { timeout: 30_000 },
  use: {
    baseURL: `https://${appDomain}`,
    ignoreHTTPSErrors: true,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'on',
    launchOptions: { args: [`--host-resolver-rules=MAP * ${deviceIp}`] },
  },
  projects: [
    {
      name: 'desktop',
      use: { ...devices['Desktop Chrome'], viewport: { width: 1440, height: 2000 } },
    },
    {
      name: 'mobile',
      use: { ...devices['Pixel 7'] },
    },
  ],
  metadata: {
    appDomain,
    fullDomain,
    artifactDir,
  },
})
