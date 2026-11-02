"""@test REL-003-AC1 REL-003-AC2: container-free release gate policy."""
import unittest
from pathlib import Path

ROOT = Path(__file__).parents[2]


class WorkflowPolicyTests(unittest.TestCase):
    def test_no_container_suite_or_runtime_workflow(self):
        self.assertFalse((ROOT / 'tests/container').exists())
        self.assertFalse((ROOT / '.github/workflows/container-checks.yml').exists())
        for path in (ROOT / '.github/workflows').glob('*.yml'):
            workflow = path.read_text()
            for forbidden in ('tests/container/', 'docker load', 'docker run', 'docker-archive:'):
                self.assertNotIn(forbidden, workflow, f'{path}: runtime test policy')

    def test_release_retains_core_validation_and_publication_order(self):
        workflow = (ROOT / '.github/workflows/release.yml').read_text()
        for required in ('scripts/check-go.sh', 'npm run build:css', 'npm run test:e2e',
                         'linux/amd64,linux/arm64', 'scripts/check-image.py',
                         'scripts/check-release-targets.py', 'scripts/release.py push',
                         'skopeo copy', '--preserve-digests', 'gh release create'):
            self.assertIn(required, workflow)
        self.assertLess(workflow.index('scripts/check-image.py'), workflow.index('scripts/release.py push'))
        self.assertLess(workflow.index('scripts/release.py push'), workflow.index('gh release create'))
