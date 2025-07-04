#!/usr/bin/env python3

import json
import subprocess
from pathlib import Path

HACK_CLOUD_CONFIG_PATH = Path("/hack/cloud-init.json")

"""Install cloud-init and launch cloud-final target on the instance"""
CLOUD_INIT_SCRIPT = """
if ! cloud-init; then
    apt-get update
    apt-get install --no-install-recommends --yes cloud-init
fi

systemctl start cloud-final.service
"""


if __name__ == "__main__":
    if not HACK_CLOUD_CONFIG_PATH.exists():
        subprocess.run(["bash", "-xe", "-c", CLOUD_INIT_SCRIPT], check=True)
        exit(0)

    hack_cloud_config = json.loads(HACK_CLOUD_CONFIG_PATH.read_text())

    for file in hack_cloud_config.get("write_files") or []:
        path = Path(file["path"])
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(file["content"])

        subprocess.run(["chmod", file["permissions"], file["path"]], check=True)
        subprocess.run(["chown", file["owner"], file["path"]], check=True)

    command = hack_cloud_config.get("runcmd") or []
    subprocess.run(["bash", "-xe", "-s"], check=True, text=True, input="\n".join(command))
