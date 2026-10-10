# Installing Reignited AP

A short, complete guide for players. The **launcher** does most of the work and
tells you what is missing, so if you get stuck follow its setup checks.

There are two halves:

1. **The world** — makes the randomizer seed. You install this if you are
   generating and hosting the game (or playing solo).
2. **The client** — connects your running game to Archipelago. Everyone who
   plays installs this.

## What you need

- **Spyro Reignited Trilogy** (PC, Steam).
- **[Archipelago](https://github.com/ArchipelagoMW/Archipelago/releases)**
  0.6.1 or newer.
- **[UE4SS](https://github.com/UE4SS-RE/RE-UE4SS/releases)** **v3.0.1** — the
  mod loader that hosts the client mod. Download the `UE4SS_v3.0.1.zip`.
- From this project's
  [Releases](https://github.com/THarrisonDev/reignited-ap/releases) page:
  - `reignited_<version>.apworld` — the world.
  - `reignited-ap-client_<version>.zip` — the launcher, the save scanner, and
    the mod.

## Part 1 — The world (host only)

Skip this if someone else is hosting and you only need to join.

1. Install Archipelago.
2. Double-click `reignited_<version>.apworld`. If you have an older
   `spyro_reignited.apworld` or `spyro1.apworld` installed, remove it first —
   they declare the same game.
3. In the Archipelago Launcher, use **Generate Template Options** to create a
   YAML, edit your settings, and drop it into the `players` folder.
4. Generate and host the seed (via `archipelago.gg` or locally). You will need
   the **server address** and your **slot name** for Part 2.

## Part 2 — The client (everyone)

### Windows

1. **Install UE4SS.** Unzip `UE4SS_v3.0.1.zip` into your game's
   `Falcon\Binaries\Win64` folder, so `dwmapi.dll` and `UE4SS.dll` sit there.
   To find the folder: Steam → right-click *Spyro Reignited Trilogy* →
   **Manage → Browse local files**, then open `Falcon\Binaries\Win64`.
2. **Extract** `reignited-ap-client_<version>.zip` somewhere convenient.
3. **Install the mod.** Copy the `srt_bridge` folder into the game's
   `Falcon\Binaries\Win64\Mods\` folder, so you have
   `...\Win64\Mods\srt_bridge\Scripts\main.lua`.
4. **Enable the mod.** Open `Falcon\Binaries\Win64\Mods\mods.txt` and make sure
   there is a line reading `srt_bridge : 1` (add it if it is missing).
5. **Install the save scanner.** Create the folder `C:\srtap_bridge\` and copy
   `srtap_checkd.exe` from the extracted client into it.
6. **Run `srtap_launcher.exe`.** It opens a page in your browser, checks your
   setup, and shows green ticks when everything is in place. Enter the server
   address and your slot name, then click **Save & launch game**.

### Linux (Steam / Proton)

The steps are the same, with one difference: the bridge folder lives inside the
game's Proton prefix, not at `C:\`. The launcher auto-detects this, but when
copying `srtap_checkd.exe` manually, use:

```
~/.local/share/Steam/steamapps/compatdata/996580/pfx/drive_c/srtap_bridge/
```

The game itself reads that same folder as `C:\srtap_bridge\`. Launch the game
through Steam so it runs under Proton.

## Playing

- From the game's main menu, pick **New Game** or **Continue**. The mod opens a
  session once you are actually playing, so you can switch or delete saves
  first. **Returning to the main menu disconnects.**
- Keep the launcher page open to watch checks you send and items you receive.
- Detailed logs are written to `C:\srtap_bridge\bridge.log` (or the prefix path
  above on Linux).

## Troubleshooting

The launcher's **Setup check** section is the first place to look — each red
item explains exactly what to fix. Common ones:

- **Check daemon missing** — `srtap_checkd.exe` is not in the bridge folder
  (step 5).
- **`lua-apclientpp.dll` missing** — it must sit next to `main.lua` in
  `Mods\srt_bridge\Scripts\`. It ships inside `srt_bridge.zip`; if you copied the
  folder by hand, make sure it came along.
- **srt_bridge not enabled** — add `srt_bridge : 1` to `Mods\mods.txt`.
- **Game folder not found** — your Steam library is probably on another drive.
  Set the **Game folder** override on the launcher page to that install's
  `Falcon\Binaries\Win64`.
- **Still not connecting** — open `bridge.log` and look at the last lines.
