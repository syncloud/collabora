# CI checks

amd64 is the primary pipeline — always check all steps on it (including test-ui and test-upgrade). arm64 runs a subset without UI tests.

# Debugging CI failures

When a CI build fails, always start by identifying the failing step:
```
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/collabora/builds/{N}" | python3 -c "
import json,sys
b=json.load(sys.stdin)
for stage in b.get('stages',[]):
    for step in stage.get('steps',[]):
        if step.get('status') == 'failure':
            print(step.get('name'), '-', step.get('status'))
"
```

Then get the step log (stage=pipeline index, step=step number):
```
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/collabora/builds/{N}/logs/{stage}/{step}" | python3 -c "
import json,sys; [print(l.get('out',''), end='') for l in json.load(sys.stdin)]
" | tail -80
```

# CI

http://ci.syncloud.org:8080/syncloud/collabora

CI is Drone CI (JS SPA). Check builds via API:
```
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/collabora/builds?limit=5"
```

Each build contains multiple pipelines (one per arch: amd64, arm64). To check status, look inside `stages` for each pipeline:
```
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/collabora/builds/{N}" | python3 -c "
import json,sys
b=json.load(sys.stdin)
for s in b.get('stages',[]):
    print(f\"{s['name']}: {s['status']}\")
"
```

## CI Artifacts

Artifacts are served at `http://ci.syncloud.org:8081` (returns JSON directory listings).

Browse the top level for a build (returns distro subdirs + snap file):
```
curl -s "http://ci.syncloud.org:8081/files/collabora/{build}-{arch}/"
```

Each distro dir contains `app/`, `platform/`, and for upgrade/UI tests also `desktop/`, `refresh.journalctl.log`, `video.mkv`:
```
curl -s "http://ci.syncloud.org:8081/files/collabora/{build}-{arch}/{distro}/"
curl -s "http://ci.syncloud.org:8081/files/collabora/{build}-{arch}/{distro}/app/"
curl -s "http://ci.syncloud.org:8081/files/collabora/{build}-{arch}/{distro}/desktop/"
```

Directory structure:
```
{build}-{arch}/
  {distro}/
    app/
      journalctl.log          # full journal from integration test teardown
      ps.log, netstat.log     # process/network state at teardown
    platform/                 # platform logs
    desktop/                  # UI test artifacts (amd64 only)
      journalctl.log
      screenshot/
        {test-name}.png
        {test-name}.html.log
      log/
    refresh.journalctl.log    # full journal from upgrade test (pre/post-refresh)
    video.mkv                 # selenium recording
```

Download a file directly:
```
curl -O "http://ci.syncloud.org:8081/files/collabora/282-amd64/buster/app/journalctl.log"
curl -O "http://ci.syncloud.org:8081/files/collabora/282-amd64/bookworm/desktop/journalctl.log"
```

# Running Drone builds locally

Generate `.drone.yml` from jsonnet (run from project root):
```
drone jsonnet --stdout --stream > .drone.yml
```

Run a specific pipeline with selected steps (e.g. amd64 up to `test bookworm`):
```
drone exec --pipeline amd64 --trusted \
  --include version \
  --include "build app" \
  --include "build python" \
  --include nginx \
  --include "nginx test" \
  --include build \
  --include cli \
  --include package \
  --include "test bookworm" \
  .drone.yml
```

Notes:
- `--trusted` is required for privileged/volume steps
- `--include` selects only listed steps (in pipeline order); omit to run all steps
- `drone jsonnet --stdout --stream` sends stderr to stderr (proto warnings are harmless)
