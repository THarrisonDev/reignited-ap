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
# keeps the match to the "X.Y.Z or newer/later/+" phrasing so unrelated
# versions (UE4SS, the client) do not trip it.
PROSE = (
    ROOT / "INSTALL.md",
    ROOT / "bundle/START-HERE.txt",
    ROOT / "world/docs/setup/en/setup_en.md",
    ROOT / "CONTRIBUTING.md",
)
VERSION_RE = re.compile(r"(\d+\.\d+\.\d+)(?= or newer| or later|\+)")


def main() -> None:
    """Compare every prose restatement against the manifest."""
    minimum = json.loads(MANIFEST.read_text())["minimum_ap_version"]
    drift = []
    for path in PROSE:
        found = VERSION_RE.findall(path.read_text())
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
