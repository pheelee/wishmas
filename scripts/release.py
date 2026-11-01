#!/usr/bin/env python3
"""Local version/ref transaction. Network writes happen only in the explicit push command.
@impl REL-001-AC1 REL-001-AC3 REL-001-AC5
"""
import argparse
import hashlib
import json
import os
import re
import subprocess
import sys
from pathlib import Path


class ReleaseError(Exception):
    pass


def git(root, *args):
    result = subprocess.run(['git', '-C', str(root), *args], text=True, capture_output=True)
    if result.returncode:
        raise ReleaseError(result.stderr.strip() or f'git {args[0]} failed')
    return result.stdout.strip()


def version_tuple(value):
    if not re.fullmatch(r'(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)', value):
        raise ReleaseError(f'Invalid stable version: {value}')
    return tuple(map(int, value.split('.')))


def records(root):
    package = json.loads((root / 'package.json').read_text())
    lock = json.loads((root / 'package-lock.json').read_text())
    value = package['version']
    version_tuple(value)
    if lock['version'] != value or lock['packages']['']['version'] != value:
        raise ReleaseError('package.json and both package-lock.json version records must agree')
    return package, lock


def fingerprint(root):
    # Includes staged and unstaged tracked changes. Untracked source is never approved.
    if git(root, 'ls-files', '--others', '--exclude-standard'):
        raise ReleaseError('Untracked files in release source')
    return hashlib.sha256(git(root, 'diff', '--binary', 'HEAD').encode()).hexdigest()


def prepare(root, bump):
    root = Path(root)
    if bump not in ('patch', 'minor', 'major'):
        raise ReleaseError('Bump must be patch, minor or major')
    if git(root, 'branch', '--show-current') != 'main':
        raise ReleaseError('Release source must be main')
    if git(root, 'status', '--porcelain'):
        raise ReleaseError('Release preparation requires a clean worktree')
    package, lock = records(root)
    current = package['version']
    tags = git(root, 'tag', '--list', 'v*').splitlines()
    stable = []
    for tag in tags:
        if re.fullmatch(r'v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)', tag):
            stable.append((version_tuple(tag[1:]), tag))
    previous = max(stable)[1] if stable else None
    if previous:
        if previous != 'v' + current:
            raise ReleaseError('Package version must match the last stable release tag')
        git(root, 'merge-base', '--is-ancestor', previous, 'HEAD')
        for filename in ('package.json', 'package-lock.json'):
            old = json.loads(git(root, 'show', f'{previous}:{filename}'))
            if old['version'] != current:
                raise ReleaseError(f'Previous tag has inconsistent {filename}')
        parts = list(version_tuple(current))
        index = {'major': 0, 'minor': 1, 'patch': 2}[bump]
        parts[index] += 1
        for i in range(index + 1, 3):
            parts[i] = 0
        version = '.'.join(map(str, parts))
    else:
        if current != '0.1.0':
            raise ReleaseError('First release must start from package version 0.1.0')
        version = '0.1.0'
    tag = 'v' + version
    if tag in tags:
        raise ReleaseError('Release tag already exists')
    base = git(root, 'rev-parse', 'HEAD')
    package['version'] = lock['version'] = lock['packages']['']['version'] = version
    for filename, value in [('package.json', package), ('package-lock.json', lock)]:
        (root / filename).write_text(json.dumps(value, indent=2, ensure_ascii=False) + '\n')
    return {'base': base, 'previous': previous, 'version': version, 'tag': tag,
            'fingerprint': fingerprint(root)}


def commit(root, plan):
    if git(root, 'branch', '--show-current') != 'main' or git(root, 'rev-parse', 'HEAD') != plan['base']:
        raise ReleaseError('Release source changed after validation')
    if fingerprint(root) != plan['fingerprint']:
        raise ReleaseError('Release files changed after preparation/validation')
    if records(Path(root))[0]['version'] != plan['version']:
        raise ReleaseError('Release version changed')
    git(root, 'add', '--', 'package.json', 'package-lock.json')
    git(root, 'commit', '--allow-empty', '-m', f"chore(release): {plan['tag']}")
    plan['revision'] = git(root, 'rev-parse', 'HEAD')
    git(root, 'tag', plan['tag'])
    return plan


def push(root, plan):
    if git(root, 'status', '--porcelain') or git(root, 'rev-parse', 'HEAD') != plan.get('revision'):
        raise ReleaseError('Validated release commit/worktree changed')
    if git(root, 'rev-parse', f"{plan['tag']}^{{commit}}") != plan['revision']:
        raise ReleaseError('Release tag does not identify the validated commit')
    remote = git(root, 'ls-remote', 'origin', 'refs/heads/main').split()
    if not remote or remote[0] != plan['base']:
        raise ReleaseError('Remote main changed; rerun validation from its current source')
    if git(root, 'ls-remote', 'origin', f"refs/tags/{plan['tag']}"):
        raise ReleaseError('Remote release tag already exists; inspect partial publication')
    # Local release commit is a child of the exact expected main. The lease prevents
    # a concurrent change even between the remote check and this atomic push.
    git(root, 'push', '--atomic', f"--force-with-lease=refs/heads/main:{plan['base']}",
        'origin', 'HEAD:refs/heads/main', f"refs/tags/{plan['tag']}")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('command', choices=['prepare', 'commit', 'push'])
    parser.add_argument('--plan', type=Path, required=True)
    parser.add_argument('--bump', choices=['patch', 'minor', 'major'])
    args = parser.parse_args()
    root = Path(git(Path.cwd(), 'rev-parse', '--show-toplevel'))
    if args.command == 'prepare':
        plan = prepare(root, args.bump)
    else:
        plan = json.loads(args.plan.read_text())
        if args.command == 'commit': commit(root, plan)
        else: push(root, plan)
    args.plan.write_text(json.dumps(plan, indent=2) + '\n')
    if os.environ.get('GITHUB_OUTPUT'):
        with open(os.environ['GITHUB_OUTPUT'], 'a') as output:
            for key in ('version', 'tag', 'base', 'revision'):
                if key in plan: output.write(f'{key}={plan[key]}\n')
    print(json.dumps(plan, indent=2))


if __name__ == '__main__':
    try: main()
    except (ReleaseError, KeyError, ValueError, OSError) as error:
        print(f'Release aborted: {error}', file=sys.stderr)
        sys.exit(1)
