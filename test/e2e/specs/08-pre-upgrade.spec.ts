import { test } from '@playwright/test'
import { ssh } from '../helpers/ssh'

const upgradeDocument = 'upgrade.txt'
const upgradePhrase = 'SurvivesTheUpgrade'

test('seed a document that must survive the upgrade', async () => {
  ssh(`sh -c 'mkdir -p /data/collabora/files'`)
  ssh(`sh -c "printf '${upgradePhrase}' > /data/collabora/files/${upgradeDocument}"`)
  ssh(`sh -c 'chown -R collabora:collabora /data/collabora/files'`)
  ssh(`sh -c 'cat /data/collabora/files/${upgradeDocument}'`)
})
