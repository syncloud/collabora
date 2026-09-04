import os
import pytest
import requests
from os.path import join
from requests.packages.urllib3.exceptions import InsecureRequestWarning
from subprocess import check_output, run
from syncloudlib.integration.hosts import add_host_alias
from syncloudlib.integration.installer import local_install

TMP_DIR = '/tmp/syncloud'

requests.packages.urllib3.disable_warnings(InsecureRequestWarning)


@pytest.fixture(scope="session")
def module_setup(request, device, artifact_dir):
    def module_teardown():
        device.run_ssh('journalctl > {0}/refresh.journalctl.log'.format(TMP_DIR), throw=False)
        device.run_ssh('ls -la /data/collabora/files > {0}/files.ls.log'.format(TMP_DIR), throw=False)
        device.scp_from_device('{0}/*'.format(TMP_DIR), artifact_dir)
        run('cp /videos/* {0}'.format(artifact_dir), shell=True)
        check_output('chmod -R a+r {0}'.format(artifact_dir), shell=True)

    request.addfinalizer(module_teardown)


def test_start(module_setup, device, device_host, app, domain):
    add_host_alias(app, device_host, domain)
    device.activated()
    device.run_ssh('mkdir -p {0}'.format(TMP_DIR), throw=False)


def test_seeded_document_exists_before_upgrade(device):
    device.run_ssh('test -f /data/collabora/files/upgrade.txt')


def test_upgrade(device_host, device_password, app_archive_path):
    local_install(device_host, device_password, app_archive_path)


def test_seeded_document_survived(device):
    device.run_ssh('test -f /data/collabora/files/upgrade.txt')
    content = device.run_ssh('cat /data/collabora/files/upgrade.txt')
    assert 'SurvivesTheUpgrade' in content, content


def test_legacy_common_dirs_removed(device):
    listing = device.run_ssh("sh -c 'ls -A /var/snap/collabora/common'")
    entries = [line for line in listing.split() if line]
    assert entries == ['web.socket'], listing


@pytest.mark.flaky(retries=10, delay=10)
def test_capabilities_after_upgrade(app_domain):
    response = requests.get('https://{0}/hosting/capabilities'.format(app_domain), verify=False)
    assert response.status_code == 200, response.text


@pytest.mark.flaky(retries=10, delay=10)
def test_spa_after_upgrade(app_domain):
    response = requests.get('https://{0}/'.format(app_domain), verify=False)
    assert response.status_code == 200, response.text
    assert 'id="app"' in response.text, response.text
