"""Offline structure checks; Home Assistant runtime tests require an HA instance."""
import ast
import json
from pathlib import Path
import unittest

ROOT = Path(__file__).resolve().parents[2]
INTEGRATION = ROOT / "custom_components" / "hammond_modern"


class IntegrationStructureTest(unittest.TestCase):
    def test_python_syntax(self):
        for path in INTEGRATION.glob("*.py"):
            with self.subTest(path=path.name):
                ast.parse(path.read_text(encoding="utf-8"))

    def test_manifest(self):
        manifest = json.loads((INTEGRATION / "manifest.json").read_text(encoding="utf-8"))
        self.assertEqual("hammond_modern", manifest["domain"])
        self.assertTrue(manifest["config_flow"])
        self.assertIn("sensor.py", {p.name for p in INTEGRATION.glob("*.py")})
        self.assertTrue(manifest["version"])
        self.assertEqual("local_polling", manifest["iot_class"])

    def test_config_translations(self):
        english = json.loads((INTEGRATION / "strings.json").read_text(encoding="utf-8"))
        translated = json.loads((INTEGRATION / "translations" / "de.json").read_text(encoding="utf-8"))
        self.assertEqual(set(english["config"]["step"]), set(translated["config"]["step"]))
        self.assertEqual(set(english["config"]["error"]), set(translated["config"]["error"]))
        for step, fields in english["config"]["step"].items():
            self.assertEqual(set(fields["data"]), set(translated["config"]["step"][step]["data"]))

    def test_read_only_client(self):
        source = (INTEGRATION / "api.py").read_text(encoding="utf-8")
        self.assertIn("allow_redirects=False", source)
        self.assertIn("Authorization", source)
        self.assertIn("async_fetch_maintenance", source)
        self.assertNotIn("session.post(", source)


if __name__ == "__main__":
    unittest.main()
