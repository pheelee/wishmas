"""@test REL-001-AC3 REL-001-AC5: publication preflight, no network requests."""
import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location('targets', Path(__file__).parents[2] / 'scripts/check-release-targets.py')
targets = importlib.util.module_from_spec(spec)
spec.loader.exec_module(targets)


class TargetsTests(unittest.TestCase):
    def test_new_package(self):
        targets.check('pheelee/wishmas', '0.1.0', lambda _: None)

    def test_existing_release(self):
        with self.assertRaises(ValueError):
            targets.check('pheelee/wishmas', '0.1.0', lambda _: {'id': 1})

    def test_existing_image_on_later_page(self):
        def read(path):
            if '/releases/' in path: return None
            if path.endswith('page=1'): return [{}] * 100
            return [{'metadata': {'container': {'tags': ['0.1.0', 'latest']}}}]
        with self.assertRaises(ValueError): targets.check('pheelee/wishmas', '0.1.0', read)

    def test_permission_failure_is_not_treated_as_missing(self):
        def read(_): raise PermissionError('not authorized')
        with self.assertRaises(PermissionError): targets.check('pheelee/wishmas', '0.1.0', read)

    def test_incomplete_previous_release_blocks_next_bump(self):
        with self.assertRaises(ValueError):
            targets.check('pheelee/wishmas', '0.1.1', lambda _: None, previous='v0.1.0')

    def test_published_previous_release_allows_new_version(self):
        def read(path):
            if path.endswith('/releases/tags/v0.1.0'): return {'draft': False, 'prerelease': False}
            if '/releases/' in path: return None
            return [{'metadata': {'container': {'tags': ['0.1.0']}}}]
        targets.check('pheelee/wishmas', '0.1.1', read, previous='v0.1.0')

    def test_published_release_without_image_blocks_next_bump(self):
        def read(path):
            if path.endswith('/releases/tags/v0.1.0'): return {'draft': False, 'prerelease': False}
            return None
        with self.assertRaises(ValueError):
            targets.check('pheelee/wishmas', '0.1.1', read, previous='v0.1.0')
