import { promises as dns } from 'node:dns'
import * as fs from 'node:fs'
import { required } from './helpers/env'

export default async function () {
  const fullDomain = required('PLAYWRIGHT_FULL_DOMAIN')
  const appDomain = required('PLAYWRIGHT_APP_DOMAIN')
  const deviceHost = required('PLAYWRIGHT_DEVICE_HOST')

  const { address } = await dns.lookup(deviceHost, { family: 4 })
  const entries = [
    `${address} ${fullDomain}`,
    `${address} ${appDomain}`,
    `${address} auth.${fullDomain}`,
  ]
  fs.appendFileSync('/etc/hosts', entries.join('\n') + '\n')

  console.log(`globalSetup: ${deviceHost} -> ${address}`)
}
