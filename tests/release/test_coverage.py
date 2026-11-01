"""@test COV-001-AC1 COV-001-AC2 COV-001-AC3"""
import importlib.util
import subprocess
import tempfile
import unittest
import xml.etree.ElementTree as ET
from pathlib import Path

spec = importlib.util.spec_from_file_location('coverage_badge', Path(__file__).parents[2] / 'scripts/coverage_badge.py')
coverage = importlib.util.module_from_spec(spec)
spec.loader.exec_module(coverage)


class CoverageTests(unittest.TestCase):
    def test_weighted_statements_not_block_average(self):
        self.assertEqual(coverage.percentage('mode: atomic\na.go:1.1,2.1 9 2\na.go:3.1,4.1 1 0\n'), '90.0')
        self.assertEqual(coverage.percentage('mode: atomic\na.go:1.1,2.1 1 0\n'), '0.0')

    def test_invalid_profiles_fail(self):
        for profile in ('', 'mode: atomic\n', 'mode: surprise\na 1 1', 'mode: atomic\nbad',
                        'mode: atomic\na.go:1.1,2.1 -1 1', 'mode: atomic\na.go:1.1,2.1 1 -1'):
            with self.subTest(profile=profile), self.assertRaises(ValueError):
                coverage.percentage(profile)

    def test_real_cli_generates_accessible_svg_and_source_report(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            profile = root / 'coverage.out'
            profile.write_text('mode: atomic\na.go:1.1,2.1 9 1\na.go:3.1,4.1 1 0\n')
            subprocess.run(['python3', str(Path(coverage.__file__)), str(profile), str(root / 'site'), 'a' * 40], check=True)
            svg = ET.parse(root / 'site/coverage.svg').getroot()
            self.assertIn('Go coverage: 90.0%', svg.attrib['aria-label'])
            report = (root / 'site/coverage.md').read_text()
            self.assertIn('90.0%', report)
            self.assertIn('a' * 40, report)
            self.assertIn('frontend', report)

    def test_first_and_subsequent_publication_never_modify_main(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            remote = root / 'remote.git'
            subprocess.run(['git', 'init', '--bare', '-q', str(remote)], check=True)
            source = root / 'source'
            source.mkdir()
            for args in (('init', '-q', '-b', 'main'), ('config', 'user.name', 'Fixture'),
                         ('config', 'user.email', 'fixture@example.invalid'),
                         ('commit', '-q', '--allow-empty', '-m', 'initial'),
                         ('push', str(remote), 'main')):
                subprocess.run(['git', '-C', str(source), *args], check=True)
            before = subprocess.check_output(['git', '--git-dir', str(remote), 'rev-parse', 'main'])
            site = root / 'site'
            site.mkdir()
            (site / 'coverage.svg').write_text('first')
            (site / 'coverage.md').write_text('report')
            coverage.publish(site, str(remote))
            self.assertEqual(subprocess.check_output(['git', '--git-dir', str(remote), 'show', 'badges:coverage.svg'], text=True), 'first')
            first = subprocess.check_output(['git', '--git-dir', str(remote), 'rev-parse', 'badges'], text=True).strip()
            coverage.publish(site, str(remote))
            self.assertEqual(subprocess.check_output(['git', '--git-dir', str(remote), 'rev-parse', 'badges'], text=True).strip(), first)
            (site / 'coverage.svg').write_text('second')
            coverage.publish(site, str(remote))
            parent = subprocess.check_output(['git', '--git-dir', str(remote), 'rev-parse', 'badges^'], text=True).strip()
            self.assertEqual(parent, first)
            self.assertEqual(subprocess.check_output(['git', '--git-dir', str(remote), 'rev-parse', 'main']), before)
            # A server rejection must fail publication and preserve the previous badge.
            hook = remote / 'hooks/pre-receive'
            hook.write_text('#!/bin/sh\nexit 1\n')
            hook.chmod(0o755)
            (site / 'coverage.svg').write_text('rejected')
            with self.assertRaises(subprocess.CalledProcessError):
                coverage.publish(site, str(remote))
            self.assertEqual(subprocess.check_output(['git', '--git-dir', str(remote), 'show', 'badges:coverage.svg'], text=True), 'second')


    def test_bad_revision_cannot_inject_report_or_svg(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            with self.assertRaises(ValueError):
                coverage.generate(root / 'missing', root / 'site', '<script>')
            self.assertFalse((root / 'site').exists())
