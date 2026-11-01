"""Behavioral release tests: all Git writes target disposable local repositories.
@test REL-001-AC1 REL-001-AC3 REL-001-AC5
"""
import importlib.util
import json
import subprocess
import tempfile
import unittest
from pathlib import Path

spec = importlib.util.spec_from_file_location('release', Path(__file__).parents[2] / 'scripts/release.py')
release = importlib.util.module_from_spec(spec)
spec.loader.exec_module(release)


class ReleaseTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name) / 'work'
        self.root.mkdir()
        self.remote = Path(self.temp.name) / 'remote.git'
        subprocess.run(['git', 'init', '--bare', '-q', str(self.remote)], check=True)
        self.git('init', '-q', '-b', 'main')
        self.git('config', 'user.name', 'Release Fixture')
        self.git('config', 'user.email', 'fixture@example.invalid')
        self.write_version('0.1.0')
        self.git('add', '.')
        self.git('commit', '-qm', 'chore: initial fixture')
        self.git('remote', 'add', 'origin', str(self.remote))
        self.git('push', '-q', 'origin', 'main')

    def git(self, *args):
        return subprocess.check_output(['git', '-C', str(self.root), *args], text=True, stderr=subprocess.PIPE).strip()

    def write_version(self, version):
        (self.root / 'package.json').write_text(json.dumps({'name': 'fixture', 'version': version}) + '\n')
        (self.root / 'package-lock.json').write_text(json.dumps({'version': version, 'lockfileVersion': 3, 'packages': {'': {'version': version}}}) + '\n')

    def previous(self, version='0.1.0'):
        self.write_version(version)
        self.git('add', '.')
        self.git('commit', '--allow-empty', '-qm', 'chore: previous release')
        self.git('tag', 'v' + version)
        self.git('push', '-q', 'origin', 'main', '--tags')

    def test_first_release_ignores_bump_and_publishes_consistent_refs(self):
        plan = release.prepare(self.root, 'major')
        self.assertEqual(plan['version'], '0.1.0')
        release.commit(self.root, plan)
        self.assertEqual(self.git('log', '-1', '--format=%s'), 'chore(release): v0.1.0')
        release.push(self.root, plan)
        self.assertIn(self.git('rev-parse', 'HEAD'), self.git('ls-remote', 'origin', 'refs/tags/v0.1.0'))
        with self.assertRaises(release.ReleaseError):
            release.push(self.root, plan)

    def test_selected_bumps_reset_components_and_update_all_records(self):
        self.previous('0.2.3')
        for bump, expected in [('patch', '0.2.4'), ('minor', '0.3.0'), ('major', '1.0.0')]:
            with self.subTest(bump=bump):
                plan = release.prepare(self.root, bump)
                self.assertEqual(plan['version'], expected)
                self.assertEqual(json.loads((self.root / 'package.json').read_text())['version'], expected)
                lock = json.loads((self.root / 'package-lock.json').read_text())
                self.assertEqual(lock['version'], expected)
                self.assertEqual(lock['packages']['']['version'], expected)
                self.git('reset', '--hard', 'HEAD')

    def test_reject_invalid_selection_branch_and_dirty_source(self):
        for bump in ['latest', '../bad', 'patch;echo bad']:
            with self.assertRaises(release.ReleaseError): release.prepare(self.root, bump)
        self.git('checkout', '-qb', 'other')
        with self.assertRaises(release.ReleaseError): release.prepare(self.root, 'patch')
        self.git('checkout', '-q', 'main')
        (self.root / 'untracked').write_text('not a tested release')
        with self.assertRaises(release.ReleaseError): release.prepare(self.root, 'patch')

    def test_inconsistent_package_or_previous_tag_fails(self):
        lock = json.loads((self.root / 'package-lock.json').read_text())
        lock['version'] = '9.0.0'
        (self.root / 'package-lock.json').write_text(json.dumps(lock))
        self.git('add', '.')
        self.git('commit', '-qm', 'chore: inconsistent fixture')
        with self.assertRaises(release.ReleaseError): release.prepare(self.root, 'patch')
        self.write_version('0.1.0')
        self.git('add', '.')
        self.git('commit', '-qm', 'chore: repair fixture')
        self.git('tag', 'v9.0.0')
        with self.assertRaises(release.ReleaseError): release.prepare(self.root, 'patch')

    def test_changes_after_prepare_cannot_be_committed_as_tested(self):
        plan = release.prepare(self.root, 'patch')
        (self.root / 'package.json').write_text('{}')
        with self.assertRaises(release.ReleaseError): release.commit(self.root, plan)
        self.assertEqual(self.git('tag', '--list'), '')

    def test_remote_main_changes_block_publication(self):
        plan = release.prepare(self.root, 'patch')
        release.commit(self.root, plan)
        other = Path(self.temp.name) / 'other'
        subprocess.run(['git', 'clone', '-qb', 'main', str(self.remote), str(other)], check=True)
        for key, value in [('user.name', 'Other'), ('user.email', 'other@example.invalid')]:
            subprocess.run(['git', '-C', str(other), 'config', key, value], check=True)
        (other / 'other.txt').write_text('concurrent main change')
        subprocess.run(['git', '-C', str(other), 'add', '.'], check=True)
        subprocess.run(['git', '-C', str(other), 'commit', '-qm', 'feat: concurrent change'], check=True)
        subprocess.run(['git', '-C', str(other), 'push', '-q'], check=True)
        with self.assertRaises(release.ReleaseError): release.push(self.root, plan)
        self.assertEqual(self.git('ls-remote', 'origin', 'refs/tags/v0.1.0'), '')


if __name__ == '__main__': unittest.main()
