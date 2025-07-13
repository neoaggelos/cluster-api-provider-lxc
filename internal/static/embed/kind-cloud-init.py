#!/usr/bin/env python3

import json
import subprocess
from pathlib import Path
from time import sleep

HACK_CLOUD_CONFIG_PATH = Path("/hack/cloud-init.json")

"""Install cloud-init and launch cloud-final target on the instance"""
CLOUD_INIT_SCRIPT = """
if ! cloud-init; then
    apt-get update
    apt-get install --no-install-recommends --yes cloud-init
fi

systemctl start --no-block cloud-final.service
"""


if __name__ == "__main__":
    # wait for instance address before proceeding
    while True:
        print("Waiting for instance address...")
        try:
            p = subprocess.run(["ip", "addr", "show"], capture_output=True, text=True)

            # wait for addresses that are "scope global" (ignore link, localhost)
            if any("scope global" in line for line in p.stdout.split("\n")):
                break
        except subprocess.CalledProcessError:
            print(f"Failed to check instance address {p.returncode=} {p.stdout=} {p.stderr=}")

        sleep(1)

    # if /hack/cloud-init.json does not exist, install cloud-init
    if not HACK_CLOUD_CONFIG_PATH.exists():
        subprocess.run(["bash", "-xe", "-c", CLOUD_INIT_SCRIPT], check=True)
        exit(0)

    # load /hack/cloud-init.json and apply manually
    hack_cloud_config = json.loads(HACK_CLOUD_CONFIG_PATH.read_text())
    for file in hack_cloud_config.get("write_files") or []:
        path = Path(file["path"])
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(file["content"])

        subprocess.run(["chmod", file["permissions"], file["path"]], check=True)
        subprocess.run(["chown", file["owner"], file["path"]], check=True)

    command = hack_cloud_config.get("runcmd") or []
    subprocess.run(["bash", "-xe", "-s"], check=True, text=True, input="\n".join(command))
