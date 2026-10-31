# Contributing to Reignited AP

Thanks for your interest in improving the Spyro Reignited Trilogy
randomizer! This guide covers the repo layout, dev setup, and how to get
your changes in.

## Repository layout

This is a parent repo holding three submodules, each independently versioned
and tagged:

- [`world/`](world) (`reignited-ap-world`) — the Archipelago world: logic,
  items, locations, and options for all three games. Python.
- [`client/`](client) (`reignited-ap-client`) — everything running on the
  player's machine: the UE4SS Lua mod (`srt_bridge`), the `checkd` Go
  daemon, and the launcher. See its
  [README](client/README.md) for component details.
- [`poptracker-pack/`](poptracker-pack) (`reignited-ap-poptracker`) — the
  PopTracker map pack and its generator.

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

`checkd` has an extensive suite backed by synthetic save fixtures in
`client/checkd/testdata` (gems, orbs, skillpoints, Sparx, bosses, records,
100% runs), generated with `pygvas` by `checkd/scripts/gen_fixtures.py` — no
real player saves. If your change affects save scanning, extend the golden
files rather than weakening assertions.

For the Lua mod, `client/.luarc.json` configures LuaLS editor support
(Lua 5.4 + the vendored `lua-apclientpp` type definitions in
`ue4ss_mod/api/`). The `api/` directory never ships to the game — only
`srt_bridge/Scripts/` does.

Testing a mod change live: drop `ue4ss_mod/srt_bridge/Scripts/` into your
game's `Mods/` directory and boot through the launcher.

The mod specs share `client/tests/lua/support/`: `harness.lua` (fake clock,
game-thread queue, fixture, PASS/FAIL runners), `reconcile.lua` and
`level_locks_suite.lua` (the shared scenario suites), and the actor fakes
(`actors.lua`, `portals.lua`, `fanfare.lua`). A new spec requires `harness` and
builds its own fakes from those pieces — a copy-pasted fake is what the
duplication gate catches.

### Linting, complexity and duplication

`make quality` at the repo root runs every linter and scan across the
submodules in one pass, each with the toolchain of the repo that owns the code:

| Scope | Tool | Catches |
| --- | --- | --- |
| `world/` | `ruff` | undefined names, unused imports, bug patterns, complexity > 10 |
| `world/` | `vulture` | functions, classes and constants nothing references |
| `client/` (Go) | `golangci-lint` (`unused`, `unparam`, `cyclop`, `staticcheck`, …) | unreachable Go decls, unused fields, dead params, complexity > 10 |
| `client/` (Lua) | `luacheck` | unused locals/arguments, dead branches, complexity > 10 |
| `client/` | `audit_dead_lua.py` | module functions no script or spec calls |
| `spyrosim/` | `golangci-lint` | unreachable simulator helpers, complexity > 10 |
| `poptracker-pack/` | `gen_pack.py --check` | pack files that drifted from the world data |
| all | `jscpd` (`.jscpd.json`) | copy-paste: any clone of 5+ lines / 50+ tokens |

The bar is hard, not advisory: **no function above cyclomatic complexity 10,
and zero duplicated blocks**. Generated data (`regions.lua`, `locations.json`,
the pack maps) is excluded from the duplication scan because the generators
emit it.

Install the tools once:

```sh
pip install ruff vulture
luarocks install luacheck
npm install -g jscpd
# golangci-lint: https://golangci-lint.run/docs/welcome/install/
```

A Lua report of `test-only (kept for the specs)` is informational — those are
module methods the specs pin. A `dead` entry fails the target: delete it or
wire it up.

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