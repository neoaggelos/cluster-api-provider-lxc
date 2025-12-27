#!/usr/bin/env python3

import argparse
import hashlib
import json
import logging
import textwrap
import urllib.request
from pathlib import Path

LOG = logging.getLogger(__name__)

STREAMS_INDEX = "streams/v1/index.json"
IMAGES_JSON = "streams/v1/images.json"
HASH_FIELD = "sha256"

STAGING_SERVER = "https://images-stg.capn.open-cloud.xyz/capn/staging/"


def fetch_url(url: str) -> bytes:
    with urllib.request.urlopen(url) as r:
        return r.read()


def sha256_file(path: str) -> str:
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(4 * 1024 * 1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def write_if_changed(path: str, data: bytes) -> bool:
    if path.exists() and path.read_bytes() == data:
        LOG.info("No changes in %s", path)
        return False

    LOG.info("Writing %s", path)
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_bytes(data)
    return True


def download_file(url: str, dest: Path, expected_hash: str):
    dest.parent.mkdir(parents=True, exist_ok=True)

    LOG.info("Downloading file %s", url)
    with urllib.request.urlopen(url) as r, open(dest, "wb") as f:
        w = 0
        while True:
            LOG.debug(
                "Downloading file %s (%.2f%%)", url, 100 * w / int(r.headers["content-length"])
            )
            chunk = r.read(4 * 1024 * 1024)
            if not chunk:
                break
            f.write(chunk)
            w += len(chunk)

    LOG.info("Downloaded file %s", url)
    actual = sha256_file(dest)
    if actual != expected_hash:
        dest.unlink(missing_ok=True)
        LOG.error("Hash mismatch for %s (expected %s, got %s)", dest, expected_hash, actual)


def sync_simplestreams(source_root: str, dest_root: str):
    source_root = source_root.rstrip("/")
    dest_root = Path(dest_root)

    LOG.info("Fetching metadata from %s", source_root)

    # Fetch remote index.json and images.json
    remote_index_json = fetch_url(f"{source_root}/{STREAMS_INDEX}")
    remote_images_json = fetch_url(f"{source_root}/{IMAGES_JSON}")

    LOG.info("Parsing image metadata from streams/v1/images.json")

    remote_images = json.loads(remote_images_json)

    # dict[<path>] = <expected hash>
    images = {
        item["path"]: item["sha256"]
        for products in remote_images.get("products", {}).values()
        for items in products.get("versions", {}).values()
        for item in items.get("items", {}).values()
    }

    LOG.info(
        "Mirroring simplestreams server with %d files: <\n%s\n>",
        len(images),
        textwrap.indent("\n".join(f"{k} -> {v}" for k, v in images.items()), "  ", lambda _: True),
    )

    for path, expected_hash in images.items():
        local_path = dest_root / path

        # Check existing file hash
        if local_path.exists():
            local_hash = sha256_file(local_path)
            if local_hash == expected_hash:
                LOG.info("File %s already exists with hash %s", path, expected_hash)
                continue
            else:
                LOG.info("Hash mismatch for %s (%s instead of %s)", path, local_hash, expected_hash)

        download_file(f"{source_root}/{path}", local_path, expected_hash)

    LOG.info("Writing image metadata")
    write_if_changed(dest_root / STREAMS_INDEX, remote_index_json)
    write_if_changed(dest_root / IMAGES_JSON, remote_images_json)

    LOG.info("Mirroring completed successfully")


def main():
    parser = argparse.ArgumentParser("mirror-simplestreams")
    parser.add_argument("--source", default=STAGING_SERVER)
    parser.add_argument("--destination", required=True)
    parser.add_argument("--debug", action=argparse.BooleanOptionalAction)

    args = parser.parse_args()
    logging.basicConfig(
        format="%(asctime)s | %(levelname)-5s | %(message)s",
        level=args.debug and logging.DEBUG or logging.INFO,
    )

    sync_simplestreams(source_root=args.source, dest_root=args.destination)


if __name__ == "__main__":
    main()
