# Installing Reignited AP

A short, complete guide for players. The **launcher** does most of the work and
tells you what is missing, so if you get stuck follow its setup checks.

Everything ships in one download — `reignited-ap_<version>_bundle.zip` — with
two halves:

1. **`client/`** — connects your running game to Archipelago. Everyone who
   plays installs this.
2. **`world/`** — the `.apworld` that makes the randomizer seed. Only needed if
   you generate and host the game (or play solo).

## What you need

- **Spyro Reignited Trilogy** (PC, Steam).
- **[Archipelago](https://github.com/ArchipelagoMW/Archipelago/releases)**
  0.6.1 or newer — only to generate and host a seed.
- The latest `reignited-ap_<version>_bundle.zip` from this project's
  [Releases](https://github.com/THarrisonDev/reignited-ap/releases) page.
- **[UE4SS](https://github.com/UE4SS-RE/RE-UE4SS/releases) v3.0.1** — the mod
  loader that hosts the client mod. The installer downloads it for you; a copy
  also ships in the bundle's `UE4SS/` folder as a fallback.

## Part 1 — The world (host only)

Skip this if someone else is hosting and you only need to join.

1. Install Archipelago.
2. Add the world to Archipelago:
   - **In the launcher:** click **Install APWorld** and pick
     `world/reignited.apworld` from the extracted bundle. Dragging the file
     onto the launcher does the same thing. If the launcher says the world is
     already loaded, close and reopen it.
   - **By hand:** copy it into Archipelago's `custom_worlds` folder.
   - Double-clicking the file only works where `.apworld` files are already
     associated with the launcher, so treat it as optional.
   Keep the file name as `reignited.apworld`: Archipelago imports a world under
   its file name, so a renamed copy does not load.
   If you installed an older build of this project — `spyro_reignited.apworld`
   or a dated `reignited_<version>.apworld` — remove that one first: it declares
   the same game, and Archipelago only loads one of them.
3. In the Archipelago Launcher, use **Generate Template Options** to create a
   YAML, edit your settings, and drop it into the `players` folder.
4. Generate and host the seed (via `archipelago.gg` or locally). You will need
   the **server address** and your **slot name** for Part 2.

## Part 2 — The client (everyone)

1. **Extract** `reignited-ap_<version>_bundle.zip` somewhere convenient.
2. **Run the installer.**
   - **Windows:** double-click `client\install.bat`.
   - **Linux (Steam / Proton):** run `client/install.sh`.

   It copies the save scanner into the bridge folder, installs the mod, enables
   it in `mods.txt`, and — if you don't already have UE4SS — downloads and
   installs UE4SS v3.0.1 into the game folder. It finds Steam wherever it is
   installed (registry on Windows; native, Flatpak and Snap on Linux) and
   searches every Steam library, so a game on another drive is found too. On
   Linux it writes into the game's Proton prefix automatically. Each step
   prints a `[ok]`/`[FAIL]` line.

   If the game has **never been launched**, start it once from Steam and quit,
   then run the installer again. The game-folder steps work either way, but
   Steam only creates the Proton prefix on first launch and the save scanner
   needs it. If the game was already running, restart it so it loads the
   updated mod.
3. **Open the launcher** (`client\srtap_launcher.exe` on Windows,
   `client/srtap_launcher` on Linux). It opens a page in your browser and shows
   green ticks when everything is in place; the **Install / update setup**
   button re-runs the installer any time. Enter the server address and your slot
   name, then click **Save & launch game**.

   The launcher re-detects the folders every time it starts (the fields say
   *(auto-detected)*). Type in a folder only to override that guess; **↺ Reset to
   auto-detected** puts it back. If Steam or the game moves, press Reset rather
   than fixing both fields by hand.

   **Linux:** add this to the game's Steam launch options (right-click the game
   → Properties → Launch Options):

   ```
   WINEDLLOVERRIDES="dwmapi=n,b" %command%
   ```

   Without it the game runs but the mod never loads.

### If the installer is blocked (manual install)

Same result as the installer, done by hand: three unzips or copies and two short
text edits. Nothing is downloaded and nothing needs admin rights.

Only two places come into it, so the steps below just call them by name:

- **Game folder** — the one holding `Falcon.exe`. In Steam: right-click the game
  → **Manage → Browse local files**, then open `Falcon\Binaries\Win64`.
- **Bridge folder** — where the mod keeps its settings and log:
  `%LOCALAPPDATA%\srtap_bridge` on Windows (usually
  `C:\Users\<you>\AppData\Local\srtap_bridge`), or that same path inside the
  game's Proton prefix on Linux. An older install that used `C:\srtap_bridge`
  keeps working: running the installer once copies its files to the new
  location.

1. **Copy in UE4SS.** Unzip `UE4SS/UE4SS_v3.0.1.zip` from the bundle into the game
   folder, so `dwmapi.dll` and `UE4SS.dll` sit beside the game's `.exe`. Use
   exactly this version; the mod is built against it.

2. **Copy in the mod.** Extract `client/srt_bridge.zip` into the game folder's
   `Mods` folder — you want a single `srt_bridge` folder inside `Mods`, holding
   `Scripts\main.lua` and `Scripts\lua-apclientpp.dll`. (Extracting *into* a
   folder already named `srt_bridge` buries it one level too deep.)

3. **Turn the mod on.** Add this line on its own to `mods.txt` in that same
   `Mods` folder, and save:

   ```
   srt_bridge : 1
   ```

   No `;` in front — that comments it out.

4. **Copy in the save scanner.** Put `client/srtap_checkd.exe` in the bridge
   folder. That is what reads your save and reports checks; the mod starts it
   for you.

5. **Type in your server details.** The easy way is to run the launcher once and
   press **Save & launch game** — the installer is the only part you skipped. If
   you cannot run it at all, create `config.txt` in the bridge folder with:

   ```
   host=archipelago.gg:38281
   slot=YourSlotName
   password=
   spawn_checkd=1
   ```

   `host` is your server's address (no `ws://` needed), `slot` is your slot name,
   and `spawn_checkd=1` lets the mod start the scanner on its own.

6. **Start the game from Steam** — always via Steam, since that is what runs it
   under Proton on Linux. Pick a save and play; the mod connects once you are in
   a level.

   On Linux, also add the launch options from Part 2.

To check your work, open the launcher: its **Setup check** list green-ticks
exactly these items, with a one-line fix for any that it does not.

## Playing

- From the game's main menu, pick **New Game** or **Continue**. The mod opens a
  session once you are actually playing, so you can switch or delete saves
  first. **Returning to the main menu disconnects.**
- Keep the launcher page open to watch checks you send and items you receive.
- Detailed logs are written to `bridge.log` in the bridge folder.

## Optional — PopTracker

The bundle also includes an optional PopTracker pack in `poptracker/` for a
desktop tracker window. You do not need it to play. Install
[PopTracker](https://github.com/black-sliver/PopTracker/releases) (v0.31.0 or
newer), copy the `spyro_reignited_ap` pack into PopTracker's `packs` folder, and
pick the **Archipelago** variant. See `poptracker/README.txt`.

## Troubleshooting

The launcher's **Setup check** section is the first place to look — each red
item explains exactly what to fix. Common ones:

- **Check daemon missing** — `srtap_checkd.exe` is not in the bridge folder.
  Click **Install / update setup**, or copy it in by hand (Part 2).
- **`lua-apclientpp.dll` missing** — it must sit next to `main.lua` in the mod's
  `Scripts` folder. It ships inside `srt_bridge.zip`; if you copied the folder by
  hand, make sure it came along.
- **srt_bridge not enabled** — add `srt_bridge : 1` to `mods.txt`.
- **Game folder not found** — the launcher searches every Steam library the
  client knows about. If it still misses, open **Folder locations**, check the
  discovery lines against your install, and set the **Game folder** override to
  the folder holding `Falcon.exe` on that drive; then press **Save config**.
  Moving a library later is handled by **↺ Reset to auto-detected**.
- **No Proton prefix (Linux)** — the game has not been launched yet. Start it
  once from Steam, quit, then click **Install / update setup** again.
- **Game runs but the mod never connects (Linux)** — the launch options are
  missing. Add `WINEDLLOVERRIDES="dwmapi=n,b" %command%` to the game's Steam
  launch options; without it the game runs but nothing loads, and there is no
  `UE4SS.log` to look at.
- **Bridge folder looks wrong after moving the game** — press **↺ Reset to
  auto-detected**. The launcher prefers the prefix the game last wrote to, so a
  stale folder should no longer win.
- **Still not connecting** — open `bridge.log` in the bridge folder and look at
  the last lines, or send the logs in (next item).
- **Asking for help** — click **Download logs for support** in the launcher, or
  run `collect-logs.bat` (Windows) / `collect-logs.sh` (Linux) from the client
  folder. Either one writes `srtap-logs-<date>.zip` next to itself with the
  recent part of `bridge.log`, `checkd.log` and `UE4SS.log`, the bridge files,
  and a `meta.txt` describing the install. Passwords, server addresses and slot
  names are redacted and **save files are never included**, so the zip is safe
  to attach to a bug report.

## Licenses

Reignited AP is GPL-3.0. The bundled third-party components (UE4SS,
lua-apclientpp, htmx, coder/websocket) are listed with their licenses in the
bundle's `THIRD-PARTY-NOTICES.md`.
