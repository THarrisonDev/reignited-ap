# Client path contract

Last updated: Nov 2026. The single description of where the client looks for
Steam, the game, the Proton prefix and the bridge folder. Every component that
resolves a path — the launcher and installer (`client/launcher/`), the mod
(`logger.lua`), checkd (`client/checkd/`) and the support-log collectors
(`client/install/collect-logs.*`) — implements these rules identically. If a
rule changes here, it changes in all of them.

## Bridge folder

The bridge holds `config.txt`, `bridge.log`, `checks.json`, `session.txt`,
`checkd.*` and `srtap_checkd.exe`.

| Context | New location | Legacy (fallback) |
|---|---|---|
| Windows | `%LOCALAPPDATA%\srtap_bridge` | `C:\srtap_bridge` |
| Proton | `<prefix>/drive_c/users/<user>/AppData/Local/srtap_bridge` | `<prefix>/drive_c/srtap_bridge` |

**Selection rule** (identical everywhere):

1. Candidates are `[new, legacy]` in that order.
2. Select the first candidate that contains `config.txt` or
   `srtap_checkd.exe`.
3. If neither does, select **new**.

`<user>` is a directory under `<prefix>/drive_c/users/` that contains
`AppData/Local`. Steam runs the game as `steamuser`, so that account is the
bridge location whenever it exists; a prefix without it falls back to the user
with the newest save, then the first in lexical order. (Save discovery is
separate: it always prefers the newest save, whichever profile holds it.)

**Migration**: only `Install` migrates. If the selection rule picks the legacy
folder and the new folder does not yet have `config.txt`/`srtap_checkd.exe`,
copy the whole legacy tree into the new folder (do not delete the legacy tree),
then delete `checkd.hb`, `checks.json` and `session.txt` in the new folder:
they are session-scoped and a copied heartbeat would make a fresh checkd exit.
After one install both sides select new, because new now contains the files.

The launcher never creates `config.txt` when the bridge's parent directory does
not exist yet (on Linux that would fabricate the Proton prefix), and it never
creates `pfx` at all.

## Game and Steam

- **Game folder**: `<library>/steamapps/common/Spyro Reignited Trilogy/Falcon/Binaries/Win64`.
- A library is preferred when it holds `steamapps/appmanifest_996580.acf`
  (the copy Steam considers installed) over one that merely has the folder
  (stale leftovers). Unreadable/unmounted libraries are skipped.

**Steam roots**, in order:

- Every OS: `$STEAM_ROOT` when set.
- Windows: registry `HKCU\Software\Valve\Steam` value `SteamPath`;
  `HKLM\SOFTWARE\WOW6432Node\Valve\Steam` value `InstallPath`;
  `HKLM\SOFTWARE\Valve\Steam` value `InstallPath`; then
  `%ProgramFiles(x86)%\Steam` and `%ProgramFiles%\Steam`.
- Linux: `$XDG_DATA_HOME/Steam` (default `~/.local/share/Steam`),
  `~/.steam/steam`, `~/.var/app/com.valvesoftware.Steam/.local/share/Steam`
  (Flatpak), `~/snap/steam/common/.local/share/Steam` (Snap).

**Libraries**: each root, plus every library path in
`<root>/steamapps/libraryfolders.vdf` (modern `"path" "<dir>"` and legacy
`"<n>" "<dir>"` forms, absolute paths only). Dedupe roots and libraries after
`filepath.EvalSymlinks`/`Clean`; dedupe case-insensitively on Windows.

## Proton prefix

`<library>/steamapps/compatdata/996580/pfx`, usable only when
`pfx/drive_c` exists. **Never create `pfx`** (or anything above `drive_c`):
a partial prefix makes Steam/Proton reuse a broken prefix.

**Selection**, first match wins:

1. The prefix whose bridge folder (new or legacy) has the newest `bridge.log`.
2. The prefix whose `drive_c/users/*/AppData/Local/Falcon/Saved/SaveGames`
   holds the newest `FalconSave.*.sav`.
3. The prefix in the library that holds the game.
4. The only existing prefix.
5. None: bridge-dependent steps are deferred, with the message
   "Launch the game once through Steam, then click Install again."

## Auto-detected vs overridden

`paths.json` (in the launcher's user config dir) stores:

```json
{ "config_dir": "...", "game_win64": "...",
  "config_dir_auto": true, "game_win64_auto": true }
```

- A field is auto when its flag is true. An old state file without flags is
  treated as auto when the saved value equals the current auto-detected value,
  and as an override otherwise.
- On apply, a field whose posted value differs from the value the page rendered
  becomes an override; an unchanged field keeps its current flag. Auto values
  are re-detected on every launcher start and on every action.
- Reset restores both fields to auto.
- The auto bridge folder is chosen by the prefix rules above. When the game
  folder is overridden, the auto bridge first considers the overridden game's
  own library, and still prefers a prefix with newer evidence.

## Verify

- Bridge folder: fails when the resolved folder is absent. On Linux with no
  usable prefix, the fix is "launch the game once through Steam, then Install
  again", not "create it".
- Bridge matches the game (Linux): fails when another prefix's bridge folder
  has a newer `bridge.log` than the selected one, naming that folder. Passes
  when no bridge log exists anywhere (nothing to contradict), so a fresh
  install is not a false red.

## Install

1. Refuse a missing `GameWin64` before writing anything.
2. Preflight-write both the game folder and the bridge parent; report a
   targeted fix on permission failure ("Steam is under Program Files — run
   install.bat as administrator").
3. Game-folder steps always run: UE4SS, `srt_bridge`, `mods.txt`. They do not
   need a prefix.
4. Bridge steps run only when the bridge parent exists (Windows
   `%LOCALAPPDATA%`, or an existing prefix `drive_c`), after migration. With no
   prefix, report the deferred-install message and leave the game steps green.

## checkd

- `--drop-dir` default: the bridge folder by the selection rule (legacy
  fallback), else the new path.
- `--log` default: `<bridge>/checkd.log`.
- `--save-dir` default: `%LOCALAPPDATA%\Falcon\Saved\SaveGames`; when unset,
  `<drive_c>/users/*/AppData/Local/Falcon/Saved/SaveGames` (prefer the newest
  save), then the legacy `C:/users/steamuser/...` constant.

## Mod

- `logger.lua` resolves the bridge folder by the same rule: base is
  `os.getenv("LOCALAPPDATA")` with `\` normalized to `/`, joined with
  `srtap_bridge`; legacy is the fixed `C:/srtap_bridge/` prefix. It exposes
  `M.bridge_dir` (with trailing slash).
- `checks_drop.lua`, `checkd_notice.lua` and `fillers.lua` derive their default
  paths from `logger.bridge_dir`; the host still passes explicit values.

## Support collectors

`collect-logs.sh` / `collect-logs.ps1` find the same roots, libraries, game,
bridge and save folders, including the new bridge location and the legacy one.
They must not keep a divergent copy of the library parser (legacy and modern
forms both).

## Diagnostics

`srtap_launcher -detect` prints the OS, every Steam root and library, the game,
every usable prefix with its bridge and log/save evidence times, the chosen
prefix, and the resolved folders with their auto flags. It creates nothing. The
launcher page shows the same discovery lines under Folders on the Setup tab, and
the support bundle's `meta.txt` records the resolved paths.
