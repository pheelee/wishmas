#!/usr/bin/env python3
"""Verify the actual unpublished OCI archive before loading or publishing it.
@impl REL-001-AC2 REL-001-AC3 REL-001-AC4
@test REL-001-AC2 REL-001-AC3 REL-001-AC4
"""
import hashlib
import io
import json
import struct
import sys
import tarfile


def require(condition, message):
    if not condition: raise ValueError(message)


def check(filename, version, revision):
    with tarfile.open(filename) as archive:
        def blob(descriptor):
            digest = descriptor['digest']
            algorithm, value = digest.split(':')
            require(algorithm == 'sha256', 'Unsupported OCI digest algorithm')
            data = archive.extractfile('blobs/sha256/' + value).read()
            require(hashlib.sha256(data).hexdigest() == value, 'OCI blob digest mismatch')
            return data
        root = json.load(archive.extractfile('index.json'))
        require(len(root['manifests']) == 1, 'Expected one multi-platform root index')
        index = json.loads(blob(root['manifests'][0]))
        manifests = index['manifests']
        require(len(manifests) == 2, 'Expected exactly two image manifests')
        platforms = {m['platform']['os'] + '/' + m['platform']['architecture'] for m in manifests}
        require(platforms == {'linux/amd64', 'linux/arm64'}, 'Missing or unexpected image platform')
        for item in manifests:
            arch = item['platform']['architecture']
            manifest = json.loads(blob(item))
            config = json.loads(blob(manifest['config']))
            require(config['architecture'] == arch and config['os'] == 'linux', 'Image platform mismatch')
            labels = config['config']['Labels']
            require(labels['org.opencontainers.image.version'] == version, 'Image version mismatch')
            require(labels['org.opencontainers.image.revision'] == revision, 'Image revision mismatch')
            require(labels['org.opencontainers.image.source'] == 'https://github.com/pheelee/wishmas', 'Image source mismatch')
            require(config['config']['User'] == '1000:1000', 'Image must retain non-root identity')
            found = set()
            for layer in manifest['layers']:
                with tarfile.open(fileobj=io.BytesIO(blob(layer)), mode='r:*') as files:
                    for name in ('wishmas', 'wishmas-admin'):
                        try: member = files.extractfile(name)
                        except KeyError: continue
                        require(member is not None, f'{name} is not a regular binary')
                        header = member.read(20)
                        require(header[:6] == b'\x7fELF\x02\x01', f'{name} must be 64-bit little-endian ELF')
                        require(struct.unpack('<H', header[18:20])[0] == {'amd64': 62, 'arm64': 183}[arch], f'{name} has wrong executable architecture')
                        found.add(name)
            require(found == {'wishmas', 'wishmas-admin'}, f'Missing executables for {arch}')
            print(f'PASS {arch}: manifest, blobs, labels, non-root user and both ELF architectures')


if __name__ == '__main__':
    try: check(*sys.argv[1:])
    except Exception as error:
        print(f'Image verification failed: {error}', file=sys.stderr)
        sys.exit(1)
