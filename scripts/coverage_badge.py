#!/usr/bin/env python3
"""Generate and optionally publish a Go statement coverage badge.
@impl COV-001-AC1 COV-001-AC2 COV-001-AC3
"""
import argparse
import re
import shutil
import subprocess
import tempfile
from decimal import Decimal, ROUND_HALF_UP
from pathlib import Path


def percentage(profile):
    lines = profile.splitlines()
    if not lines or lines[0] not in ('mode: atomic', 'mode: count', 'mode: set'):
        raise ValueError('Invalid Go coverage header')
    total = covered = 0
    for line in lines[1:]:
        match = re.fullmatch(r'.+:\d+\.\d+,\d+\.\d+ (\d+) (\d+)', line)
        if not match:
            raise ValueError('Invalid Go coverage block')
        statements, count = map(int, match.groups())
        total += statements
        if count:
            covered += statements
    if not total:
        raise ValueError('Coverage profile contains no statements')
    return str((Decimal(covered) * 100 / Decimal(total)).quantize(Decimal('0.1'), rounding=ROUND_HALF_UP))


def generate(profile, output, revision):
    if not re.fullmatch(r'[a-f0-9]{40}', revision):
        raise ValueError('Expected a complete Git commit SHA')
    value = percentage(profile.read_text())
    output.mkdir(parents=True, exist_ok=True)
    # All interpolated text is validated numeric/hexadecimal; no profile paths enter SVG.
    (output / 'coverage.svg').write_text(f'''<svg xmlns="http://www.w3.org/2000/svg" width="156" height="20" role="img" aria-label="Go coverage: {value}%">
<title>Go coverage: {value}%</title>
<rect width="156" height="20" rx="3" fill="#555"/>
<path d="M94 0h59q3 0 3 3v14q0 3-3 3H94z" fill="#007f69"/>
<g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,sans-serif" font-size="11">
<text x="47" y="14">Go coverage</text><text x="125" y="14">{value}%</text></g></svg>\n''')
    (output / 'coverage.md').write_text(f'''# Go statement coverage

**{value}%** of Go statements covered by the CI race-test run.

Source: [{revision}](https://github.com/pheelee/wishmas/commit/{revision}).

Scope: Go package tests using default per-package instrumentation. Calculated from all profile blocks weighted by their statement counts; the named-function total from `go tool cover -func` can differ. This excludes frontend JavaScript, browser and container test coverage. No minimum coverage threshold is enforced.
''')


def publish(site, remote):
    """Publish with ordinary pushes: concurrent/non-fast-forward updates fail safely."""
    with tempfile.TemporaryDirectory(prefix='wishmas-coverage-') as directory:
        root = Path(directory)

        def git(*args, **kwargs):
            return subprocess.run(['git', '-C', str(root), *args], check=True, text=True, **kwargs)

        git('init', '-q', '-b', 'badges')
        git('config', 'user.name', 'github-actions[bot]')
        git('config', 'user.email', '41898282+github-actions[bot]@users.noreply.github.com')
        git('remote', 'add', 'origin', remote)
        refs = git('ls-remote', '--heads', 'origin', 'refs/heads/badges', capture_output=True).stdout
        if refs.strip():
            git('fetch', '--depth=1', 'origin', 'refs/heads/badges')
            git('reset', '--hard', 'FETCH_HEAD')
        for filename in ('coverage.svg', 'coverage.md'):
            shutil.copyfile(site / filename, root / filename)
        git('add', 'coverage.svg', 'coverage.md')
        if not git('status', '--porcelain', capture_output=True).stdout.strip():
            return
        git('commit', '-qm', 'docs: update Go coverage badge')
        git('push', 'origin', 'HEAD:refs/heads/badges')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('profile', type=Path)
    parser.add_argument('output', type=Path)
    parser.add_argument('revision')
    parser.add_argument('--publish', metavar='REMOTE')
    args = parser.parse_args()
    generate(args.profile, args.output, args.revision)
    if args.publish:
        publish(args.output, args.publish)


if __name__ == '__main__':
    main()
