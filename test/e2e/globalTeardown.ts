import { ssh, scpFrom } from './helpers/ssh'
import { required } from './helpers/env'
import * as path from 'node:path'
import * as fs from 'node:fs'
import { execSync } from 'node:child_process'

const TMP_DIR = '/tmp/syncloud/collabora-ui'

export default async function () {
  const artifactRoot = required('PLAYWRIGHT_ARTIFACT_DIR')
  const project = required('PLAYWRIGHT_PROJECT')
  const out = path.join(artifactRoot, 'playwright', project)
  fs.mkdirSync(out, { recursive: true })

  ssh(`mkdir -p ${TMP_DIR}`, { throw: false })
  ssh(`journalctl > ${TMP_DIR}/journalctl.log`, { throw: false })
  ssh(`ls -la /var/snap/collabora/current/config > ${TMP_DIR}/config.ls.log 2>&1`, { throw: false })
  ssh(`cat /var/snap/collabora/current/config/coolwsd.xml > ${TMP_DIR}/coolwsd.xml 2>&1`, { throw: false })
  scpFrom(`${TMP_DIR}/*`, out, { throw: false })
  try { execSync(`chmod -R a+r ${out}`) } catch {}
}
