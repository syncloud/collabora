local name = 'collabora';
local code = '25.04.9.4.1';
local go = '1.25';
local nginx = '1.29.3-alpine3.22';
local debian = 'bookworm-slim';
local platform = '26.08.01';
local playwright = 'mcr.microsoft.com/playwright:v1.59.1-jammy';
local store_publisher = 'stable-346';
local python = '3.12-slim-bookworm';
local distro_default = 'bookworm';
local distros = ['bookworm', 'buster'];

local platform_image(distro) =
  'syncloud/platform-' + distro + ':' + platform;


local build(arch, test_ui) = [{
  kind: 'pipeline',
  type: 'docker',
  name: arch,
  platform: {
    os: 'linux',
    arch: arch,
  },
  steps: [
    {
      name: 'build app',
      image: 'collabora/code:' + code,
      user: 'root',
      commands: [
        './app/build.sh',
      ],
    },
    {
      name: 'nginx',
      image: 'nginx:' + nginx,
      commands: [
        './nginx/build.sh',
      ],
    },
    {
      name: 'nginx test',
      image: platform_image(distro_default),
      commands: [
        './nginx/test.sh',
      ],
    },
    {
      name: 'build',
      image: 'debian:' + debian,
      commands: [
        './build.sh',
      ],
    },
    {
      name: 'cli',
      image: 'golang:' + go,
      commands: [
        'cd cli',
        'CGO_ENABLED=0 go build -o ../build/snap/meta/hooks/install ./cmd/install',
        'CGO_ENABLED=0 go build -o ../build/snap/meta/hooks/configure ./cmd/configure',
        'CGO_ENABLED=0 go build -o ../build/snap/meta/hooks/pre-refresh ./cmd/pre-refresh',
        'CGO_ENABLED=0 go build -o ../build/snap/meta/hooks/post-refresh ./cmd/post-refresh',
        'CGO_ENABLED=0 go build -o ../build/snap/bin/cli ./cmd/cli',
      ],
    },
    {
      name: 'package',
      image: 'debian:' + debian,
      commands: [
        './package.sh ' + name + ' $DRONE_BUILD_NUMBER',
      ],
    },
  ] + [
    {
      name: 'test ' + distro,
      image: 'python:' + python,
      commands: [
        'cd test',
        './deps.sh',
        'py.test -x -s test.py --distro=' + distro + ' --ver=$DRONE_BUILD_NUMBER --app=' + name,
      ],
    }
    for distro in distros
  ] + (if test_ui then [
    {
      name: 'e2e',
      image: playwright,
      commands: [
        './test/e2e/run.sh e2e specs/01-login.spec.ts',
      ],
    },
    {
      name: 'e2e-mobile',
      image: playwright,
      commands: [
        './test/e2e/run.sh e2e-mobile specs/01-login.spec.ts mobile',
      ],
    },
    {
      name: 'test-upgrade',
      image: 'python:' + python,
      commands: [
        'cd test',
        './deps.sh',
        'py.test -x -s upgrade.py --distro=' + distro_default + ' --ver=$DRONE_BUILD_NUMBER --app=' + name,
      ],
      privileged: true,
    },
  ] else []) + [
    {
      name: 'publish',
      image: 'syncloud/store-publisher:' + store_publisher,
      environment: {
        SYNCLOUD_TOKEN: { from_secret: 'SYNCLOUD_TOKEN' },
      },
      command: ['snap', '-c', '${DRONE_BRANCH}'],
      when: {
        branch: ['master', 'stable'],
        event: ['push'],
      },
    },
    {
      name: 'artifact',
      image: 'appleboy/drone-scp:1.6.4',
      settings: {
        host: {
          from_secret: 'artifact_host',
        },
        username: 'artifact',
        key: {
          from_secret: 'artifact_key',
        },
        timeout: '2m',
        command_timeout: '2m',
        target: '/home/artifact/repo/' + name + '/${DRONE_BUILD_NUMBER}-' + arch,
        source: 'artifact/*',
        strip_components: 1,
      },
      when: {
        status: ['failure', 'success'],
        event: ['push'],
      },
    },
  ],
  trigger: {
    event: ['push'],
  },
  services: [
    {
      name: name + '.' + distro + '.com',
      image: platform_image(distro),
      privileged: true,
      entrypoint: ['/bin/sh', '-c', "mkdir -p /etc/systemd/system/snapd.service.d && printf '[Service]\\nExecStartPost=/bin/sh -c \"/usr/bin/snap set system refresh.hold=2099-01-01T00:00:00Z\"\\n' > /etc/systemd/system/snapd.service.d/disable-refresh.conf && exec /sbin/init"],
      volumes: [
        {
          name: 'dbus',
          path: '/var/run/dbus',
        },
        {
          name: 'dev',
          path: '/dev',
        },
      ],
    }
    for distro in distros
  ],
  volumes: [
    {
      name: 'dbus',
      host: {
        path: '/var/run/dbus',
      },
    },
    {
      name: 'dev',
      host: {
        path: '/dev',
      },
    },
  ],
}];

build('amd64', true) +
build('arm64', false)
