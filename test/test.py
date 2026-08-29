import os
import pytest
import requests
from os.path import join
from requests.packages.urllib3.exceptions import InsecureRequestWarning
from subprocess import check_output
from syncloudlib.integration.hosts import add_host_alias
from syncloudlib.integration.installer import local_install
from retrying import retry

TMP_DIR = '/tmp/syncloud'

requests.packages.urllib3.disable_warnings(InsecureRequestWarning)


@pytest.fixture(scope="session")
def module_setup(request, device, platform_data_dir, app_dir, artifact_dir):
    def module_teardown():
        platform_log_dir = join(artifact_dir, 'platform_log')
        os.mkdir(platform_log_dir)
        device.scp_from_device('{0}/log/*'.format(platform_data_dir), platform_log_dir)
        device.run_ssh('top -bn 1 -w 500 -c > {0}/top.log'.format(TMP_DIR), throw=False)
        device.run_ssh('ps auxfw > {0}/ps.log'.format(TMP_DIR), throw=False)
        device.run_ssh('ss -lntp > {0}/netstat.log'.format(TMP_DIR), throw=False)
        device.run_ssh('journalctl | tail -1000 > {0}/journalctl.log'.format(TMP_DIR), throw=False)
        device.run_ssh('journalctl -u snap.collabora.server --no-pager > {0}/server.log'.format(TMP_DIR),
                       throw=False)
        device.run_ssh('journalctl -u snap.collabora.backend --no-pager > {0}/backend.log'.format(TMP_DIR),
                       throw=False)
        device.run_ssh('ls -la /snap/collabora > {0}/snap.ls.log'.format(TMP_DIR), throw=False)
        device.run_ssh('ls -la /var/snap/collabora/current > {0}/var.snap.current.ls.log'.format(TMP_DIR),
                       throw=False)
        device.run_ssh('ls -la /var/snap/collabora/common > {0}/var.snap.common.ls.log'.format(TMP_DIR),
                       throw=False)
        device.run_ssh('ls -la /var/snap/collabora/current/config > {0}/config.ls.log'.format(TMP_DIR),
                       throw=False)
        device.run_ssh('cat /var/snap/collabora/current/config/nginx.conf > {0}/nginx.conf'.format(TMP_DIR),
                       throw=False)
        device.run_ssh('ls -la /data/collabora > {0}/data.ls.log'.format(TMP_DIR), throw=False)

        app_log_dir = join(artifact_dir, 'log')
        os.mkdir(app_log_dir)
        device.scp_from_device('{0}/*'.format(TMP_DIR), app_log_dir)
        check_output('chmod -R a+r {0}'.format(artifact_dir), shell=True)

    request.addfinalizer(module_teardown)


def settle(device):
    device.run_ssh('snap wait system seed.loaded', retries=100, throw=False)
    device.run_ssh('snap set system refresh.hold=2099-01-01T00:00:00Z', retries=20, throw=False)
    device.run_ssh('snap abort --last=auto-refresh', throw=False)
    device.run_ssh('snap watch --last=auto-refresh', throw=False)


def test_start(module_setup, device, device_host, app, domain):
    add_host_alias(app, device_host, domain)
    device.run_ssh('date', retries=10)
    device.run_ssh('mkdir {0}'.format(TMP_DIR))
    settle(device)


@retry(stop_max_attempt_number=10, wait_fixed=5000)
def activate(device):
    response = device.activate_custom()
    assert response.status_code == 200, response.text


def test_activate_device(device):
    activate(device)


def test_install(app_archive_path, device_session, device_host, device_password):
    local_install(device_host, device_password, app_archive_path)


def test_storage_change_event(device):
    device.run_ssh('snap run collabora.storage-change > {0}/storage-change.log'.format(TMP_DIR))


def test_access_change_event(device):
    device.run_ssh('snap run collabora.access-change > {0}/access-change.log'.format(TMP_DIR))


def test_access_change_regenerates_coolwsd_config(device, domain):
    config = device.run_ssh('cat /var/snap/collabora/current/config/coolwsd.xml')
    assert 'server_name' in config, config[:200]
    assert domain.replace('.', chr(92) + '.') in config, 'device domain missing from alias_groups'


def test_oidc_registered_once(device):
    device.run_ssh('test -f /var/snap/collabora/current/secret/oidc.secret')
    before = device.run_ssh('cat /var/snap/collabora/current/secret/oidc.secret').strip()
    device.run_ssh('snap run collabora.access-change')
    after = device.run_ssh('cat /var/snap/collabora/current/secret/oidc.secret').strip()
    assert before == after, 'oidc secret rotated without a url change'


def test_backend_socket(device):
    device.run_ssh('test -S /var/snap/collabora/current/run/backend.sock', retries=100)


def test_storage_dir(device):
    device.run_ssh('test -d /data/collabora/files', retries=100)


def test_admin_password_is_generated(device):
    password = device.run_ssh('cat /var/snap/collabora/current/secret/admin.password', retries=100)
    assert password.strip() != 'admin', password
    assert len(password.strip()) > 20, password


def test_snap_common_holds_only_the_web_socket(device):
    listing = device.run_ssh("sh -c 'ls -A /var/snap/collabora/common'")
    entries = [line for line in listing.split() if line]
    assert entries == ['web.socket'], listing


def test_server_started(device):
    try:
        device.run_ssh("sh -c 'ss -lnt | grep -q 127.0.0.1:9980'", retries=60)
    except Exception:
        logs = device.run_ssh('journalctl -u snap.collabora.server -n 300 --no-pager', throw=False)
        raise AssertionError('coolwsd never listened on 127.0.0.1:9980\n{0}'.format(logs))


def test_coolwsd_and_wopi_listen_on_loopback_ipv4_only(device):
    listening = device.run_ssh("sh -c 'ss -lntp'")
    for port in ['9980', '9981']:
        lines = [line for line in listening.split('\n') if ':' + port in line]
        assert lines, 'nothing listening on {0}:\n{1}'.format(port, listening)
        for line in lines:
            assert '127.0.0.1:' + port in line, line


@pytest.mark.flaky(retries=10, delay=5)
def test_capabilities(app_domain):
    response = requests.get('https://{0}/hosting/capabilities'.format(app_domain), verify=False)
    assert response.status_code == 200, response.text
    assert "productVersion" in response.text


@pytest.mark.flaky(retries=10, delay=5)
def test_discovery(app_domain):
    response = requests.get('https://{0}/hosting/discovery'.format(app_domain), verify=False)
    assert response.status_code == 200, response.text
    assert 'urlsrc' in response.text, response.text


@pytest.mark.flaky(retries=10, delay=5)
def test_spa_is_served(app_domain):
    response = requests.get('https://{0}/'.format(app_domain), verify=False)
    assert response.status_code == 200, response.text
    assert 'id="app"' in response.text, response.text


@pytest.mark.flaky(retries=10, delay=5)
def test_api_requires_a_session(app_domain):
    response = requests.get('https://{0}/api/session'.format(app_domain), verify=False)
    assert response.status_code == 401, response.text


@pytest.mark.flaky(retries=20, delay=10)
def test_oidc_login_redirects_to_authelia(app_domain):
    response = requests.get('https://{0}/oidc/start'.format(app_domain),
                            verify=False, allow_redirects=False)
    assert response.status_code == 302, response.text
    assert 'auth.' in response.headers.get('Location', ''), response.headers


@pytest.mark.flaky(retries=10, delay=5)
def test_admin_console_is_behind_sso(app_domain):
    response = requests.get('https://{0}/browser/dist/admin/admin.html'.format(app_domain),
                            verify=False, allow_redirects=False)
    assert response.status_code == 302, response.text
    assert 'auth.' in response.headers.get('Location', ''), response.headers


def test_remove(device, app):
    response = device.app_remove(app)
    assert response.status_code == 200, response.text


def test_reinstall(app_archive_path, device_host, device_password):
    local_install(device_host, device_password, app_archive_path)
