import logging
import shutil
import uuid
import re
from os.path import isfile
from os.path import join
from os.path import realpath
from subprocess import check_output, CalledProcessError

from syncloudlib import fs, linux, gen, logger
from syncloudlib.application import paths, urls, storage, service


APP_NAME = 'collabora'

USER_NAME = APP_NAME

SYSTEMD_NGINX = '{0}.nginx'.format(APP_NAME)
SYSTEMD_COLLABORA = '{0}.app'.format(APP_NAME)

class Installer:
    def __init__(self):
        if not logger.factory_instance:
            logger.init(logging.DEBUG, True)

        self.log = logger.get_logger(APP_NAME)
        self.app_dir = paths.get_app_dir(APP_NAME)
        self.common_dir = paths.get_data_dir(APP_NAME)
        self.data_dir = join('/var/snap', APP_NAME, 'current')
        self.config_dir = join(self.data_dir, 'config')

    def install_config(self):

        home_folder = join('/home', USER_NAME)
        linux.useradd(USER_NAME, home_folder=home_folder)
        storage.init_storage(APP_NAME, USER_NAME)
        templates_path = join(self.app_dir, 'config')

        variables = {
            'snap_data': self.data_dir,
            'domain': urls.get_app_domain_name(APP_NAME)
        }
        gen.generate_files(templates_path, self.config_dir, variables)

        fs.makepath(join(self.common_dir, 'log'))
        fs.makepath(join(self.common_dir, 'nginx'))
        
        cool_fileserver_path = join(self.data_dir, 'coolwsd')
        fs.makepath(cool_fileserver_path)
        fs.makepath(join(self.data_dir, 'systemplate'))
        fs.makepath(join(self.data_dir, 'child-roots'))
        shutil.copy(
            join(self.config_dir, 'discovery.xml'), 
            join(cool_fileserver_path, 'discovery.xml')
        )
        check_output('cp -r {0}/app/usr/share/coolwsd/browser {1}'
                     .format(self.app_dir, cool_fileserver_path), shell=True)
        self.fix_permissions()


    def install(self):
        self.install_config()

    def pre_refresh(self):
        pass

    def post_refresh(self):
        self.install_config()

    def configure(self):
        self.prepare_storage()
        app_storage_dir = storage.init_storage(APP_NAME, USER_NAME)

        self.on_domain_change()

        self.fix_permissions()

    def fix_permissions(self):
        check_output('chown -R {0}.{0} {1}'.format(USER_NAME, self.common_dir), shell=True)
        check_output('chown -R {0}.{0} {1}'.format(USER_NAME, self.data_dir), shell=True)

    def on_disk_change(self):
        self.prepare_storage()

    def prepare_storage(self):
        app_storage_dir = storage.init_storage(APP_NAME, USER_NAME)
        check_output('chmod 770 {0}'.format(app_storage_dir), shell=True)

    def on_domain_change(self):
        app_domain = urls.get_app_domain_name(APP_NAME)
        gen.generate_file_jinja(
            join(self.app_dir, 'config', 'code', 'coolwsd.xml'),
            join(self.data_dir, 'config', 'code', 'coolwsd.xml'),
            {'domain': app_domain}
        )
        service.restart(SYSTEMD_COLLABORA)
