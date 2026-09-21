# Contributing to Reignited AP

Thanks for your interest in improving the Spyro Reignited Trilogy
randomizer! This guide covers the repo layout, dev setup, and how to get
your changes in.

## Repository layout

This is a parent repo holding two submodules, each independently versioned
and tagged:

- [`world/`](world) (`reignited-ap-world`) — the Archipelago world: logic,
  items, locations, and options for all three games. Python.
- [`client/`](client) (`reignited-ap-client`) — everything running on the
  player's machine: the UE4SS Lua mod (`srt_bridge`), the `checkd` Go
  daemon, and the launcher. See its
  [README](client/README.md) for component details.

Start by reading the README of whichever submodule you'll work in.

## Dev setup

### Clone with submodules

```sh
git clone --recurse-submodules https://github.com/THarrisonDev/reignited-ap
cd reignited-ap
# or, if already cloned:
git submodule update --init --recursive
```

### Working on the world (Python)

The world targets a real Archipelago install. Point your checkout at your
Archipelago installation (0.6.1+) by placing/symlinking this repo's `world/`
directory as a custom world, or run generation from within an Archipelago
source checkout with `world/` on `PYTHONPATH`.

- `cd world && make check-version` — verifies the version lockstep between
  `archipelago.json` and `__init__.py` (see [RELEASE.md](world/RELEASE.md)).
- `make bump VERSION=x.y.z` — bump both version files in one step.
- Smoke test generation: install the world into an Archipelago checkout and
  generate a seed with default options.
- `.apignore` files control what gets packaged into the `.apworld` zip.

### Working on the client (Go + Lua)

Each Go component is its own module — test from within its directory:

```sh
cd client/checkd   && go test ./...
cd client/launcher && go test ./...
```

`checkd` has an extensive suite backed by captured save data in
`client/checkd/testdata` (gems, orbs, skillpoints, Sparx, bosses, records,
100% runs). If your change affects save scanning, extend the golden files
rather than weakening assertions.

For the Lua mod, a `.luarc.json` at the client root configures LuaLS editor
support (Lua 5.4 + the vendored `lua-apclientpp` type definitions in
`ue4ss_mod/api/`). The `api/` directory never ships to the game — only
`srt_bridge/Scripts/` does.

Testing a mod change live: drop `ue4ss_mod/srt_bridge/Scripts/` into your
game's `Mods/` directory and boot through the launcher.

## Submitting changes

1. Fork the repository and create a new branch for your changes.
2. Make your changes, keeping the existing code style in mind.
3. Test your changes locally:
   - World changes: `make check-version` plus a generation smoke test.
   - Client changes: `go test ./...` in each module you touched.
4. Open a pull request with a clear description of what you changed and why.

If you're not sure where to start, check the open issues for good first
tasks — or open a new issue to discuss an idea before writing code.

## Submodule notes

- Submodule PRs go to their own repos (`reignited-ap-world`,
  `reignited-ap-client`); parent-repo PRs are for README, submodule pointer
  bumps, and cross-cutting docs.
- Versions move in lockstep across the three repos — if your change bumps
  the world or client version, follow the checklist in
  [world/RELEASE.md](world/RELEASE.md).

## Licensing

By contributing, you agree your contributions are licensed under the
GPL-3.0 license that covers this project (see [LICENSE](LICENSE) in each
repo).