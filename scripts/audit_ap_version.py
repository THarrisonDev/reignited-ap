#!/usr/bin/env python3
"""Guard the Archipelago minimum version against drift across the player docs.

`world/archipelago.json` is the only machine-read copy; the player docs restate
it as prose. Nothing compares them, so bumping one and not the others would
advertise a version the apworld does not support. Exits non-zero on a mismatch
or a missing restatement.
"""

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
MANIFEST = ROOT / "world/archipelago.json"
# Every prose restatement of the minimum, in the parent checkout. The lookahead
# keeps the match to the "X.Y.Z or newer/later/+" phrasing, and OTHER_PRODUCT_RE
# drops a line that restates another product's minimum in the same phrasing.
PROSE = (
    ROOT / "INSTALL.md",
    ROOT / "bundle/START-HERE.txt",
    ROOT / "world/docs/setup/en/setup_en.md",
    ROOT / "CONTRIBUTING.md",
)
VERSION_RE = re.compile(r"(\d+\.\d+\.\d+)(?= or newer| or later|\+)")
OTHER_PRODUCT_RE = re.compile(r"Universal Tracker|UE4SS|client", re.IGNORECASE)


def restatements(path: Path) -> list[str]:
    """Return the minimum versions restated in a doc, ignoring other products."""
    versions = []
    for line in path.read_text().splitlines():
        for match in VERSION_RE.finditer(line):
            if OTHER_PRODUCT_RE.search(line[: match.start()]):
                continue
            versions.append(match.group(1))
    return versions


def main() -> None:
    """Compare every prose restatement against the manifest."""
    minimum = json.loads(MANIFEST.read_text())["minimum_ap_version"]
    drift = []
    for path in PROSE:
        found = restatements(path)
        if not found:
            drift.append(f"{path.relative_to(ROOT)}: no Archipelago version found")
            continue
        drift.extend(
            f"{path.relative_to(ROOT)}: {version} (manifest says {minimum})"
            for version in found
            if version != minimum
        )
    if drift:
        sys.exit("ERROR: Archipelago minimum version drift:\n  " + "\n  ".join(drift))
    print(f"Archipelago minimum version OK: {minimum}")


if __name__ == "__main__":
    main()
