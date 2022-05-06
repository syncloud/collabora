import subprocess

print(subprocess.check_output('snap run collabora.storage-change', shell=True))
