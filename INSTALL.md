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
2. In the extracted bundle, double-click `world/reignited_<version>.apworld`.
   If you have an older `spyro_reignited.apworld` or `spyro1.apworld` installed,
   remove it first — they declare the same game.
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
   installs UE4SS v3.0.1 into `Falcon\Binaries\Win64`. It finds the game across
   all of your Steam libraries; on Linux it writes into the Proton prefix
   automatically. Each step prints a `[ok]`/`[FAIL]` line.
3. **Open the launcher** (`client\srtap_launcher.exe` on Windows,
   `client/srtap_launcher` on Linux). It opens a page in your browser and shows
   green ticks when everything is in place; the **Install / update setup**
   button re-runs the installer any time. Enter the server address and your slot
   name, then click **Save & launch game**.

If the installer is blocked (managed machine, antivirus, no internet), do it by
hand — everything you need is in `client/` and `UE4SS/`:

- **UE4SS** — unzip `UE4SS/UE4SS_v3.0.1.zip` into `Falcon\Binaries\Win64`, so
  `dwmapi.dll` and `UE4SS.dll` sit there. To find the folder: Steam →
  right-click *Spyro Reignited Trilogy* → **Manage → Browse local files**.
- **Mod** — extract `client/srt_bridge.zip` into `...\Win64\Mods\`, so you have
  `...\Win64\Mods\srt_bridge\`, then add `srt_bridge : 1` to `Mods\mods.txt`.
- **Save scanner** — copy `client/srtap_checkd.exe` into the bridge folder:
  `C:\srtap_bridge\` on Windows,
  `~/.local/share/Steam/steamapps/compatdata/996580/pfx/drive_c/srtap_bridge/`
  on Linux (the game reads it as `C:\srtap_bridge\`).

Launch the game through Steam so it runs under Proton on Linux.

## Playing

- From the game's main menu, pick **New Game** or **Continue**. The mod opens a
  session once you are actually playing, so you can switch or delete saves
  first. **Returning to the main menu disconnects.**
- Keep the launcher page open to watch checks you send and items you receive.
- Detailed logs are written to `C:\srtap_bridge\bridge.log` (or the prefix path
  above on Linux).

## Optional — PopTracker

The bundle also includes an optional PopTracker pack in `poptracker/` for a
desktop tracker window. You do not need it to play. Install
[PopTracker](https://github.com/black-sliver/PopTracker/releases) (v0.31.0 or
newer), copy `poptracker/spyro_reignited_ap` into PopTracker's `packs` folder,
and pick the **Archipelago** variant. See `poptracker/README.txt`.

## Troubleshooting

The launcher's **Setup check** section is the first place to look — each red
item explains exactly what to fix. Common ones:

- **Check daemon missing** — `srtap_checkd.exe` is not in the bridge folder.
  Click **Install / update setup**, or copy it in by hand (Part 2).
- **`lua-apclientpp.dll` missing** — it must sit next to `main.lua` in
  `Mods\srt_bridge\Scripts\`. It ships inside `srt_bridge.zip`; if you copied the
  folder by hand, make sure it came along.
- **srt_bridge not enabled** — add `srt_bridge : 1` to `Mods\mods.txt`.
- **Game folder not found** — your Steam library is probably on another drive.
  Set the **Game folder** override on the launcher page to that install's
  `Falcon\Binaries\Win64`.
- **Still not connecting** — open `bridge.log` and look at the last lines.

## Licenses

Reignited AP is GPL-3.0. The bundled third-party components (UE4SS,
lua-apclientpp, htmx, coder/websocket) are listed with their licenses in
`client/THIRD-PARTY-NOTICES.md`.
