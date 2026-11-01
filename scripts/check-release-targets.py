#!/usr/bin/env python3
"""Fail closed on API errors; reject an existing release or immutable GHCR version.
@impl REL-001-AC3 REL-001-AC4 REL-001-AC5
"""
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request


def get(path):
    req = urllib.request.Request('https://api.github.com/' + path, headers={
        'Authorization': 'Bearer ' + os.environ['GH_TOKEN'],
        'Accept': 'application/vnd.github+json', 'X-GitHub-Api-Version': '2022-11-28'})
    try:
        with urllib.request.urlopen(req, timeout=30) as response:
            return json.load(response)
    except urllib.error.HTTPError as error:
        if error.code == 404: return None
        raise


def check(repository, version, read=get, previous=None):
    owner, name = repository.split('/')
    if previous:
        prior = read(f'repos/{repository}/releases/tags/{previous}')
        if prior is None or prior.get('draft', True) or prior.get('prerelease', True):
            raise ValueError('Previous tag has no published stable release; resolve partial publication before another bump')
    found_previous_image = not previous
    if read(f'repos/{repository}/releases/tags/v{version}') is not None:
        raise ValueError('GitHub release already exists; inspect partial publication before retrying')
    page = 1
    while True:
        versions = read(f'users/{owner}/packages/container/{urllib.parse.quote(name, safe="")}/versions?per_page=100&page={page}')
        if versions is None:
            if not found_previous_image: raise ValueError('Previous release image cannot be found; verify package access/publication')
            return  # First package publication: API reports not found.
        for item in versions:
            tags = item.get('metadata', {}).get('container', {}).get('tags', [])
            if previous and previous.removeprefix('v') in tags: found_previous_image = True
            if version in tags:
                raise ValueError('GHCR version tag already exists; never replace a published version')
        if len(versions) < 100:
            if not found_previous_image: raise ValueError('Previous release image is missing; resolve partial publication')
            return
        page += 1


if __name__ == '__main__':
    try:
        previous = json.loads(open(sys.argv[2]).read()).get('previous') if len(sys.argv) > 2 else None
        check(os.environ['GITHUB_REPOSITORY'], sys.argv[1], previous=previous)
        print('Release and immutable GHCR tag are available')
    except Exception as error:
        print(f'Release publication preflight failed: {error}', file=sys.stderr)
        sys.exit(1)
