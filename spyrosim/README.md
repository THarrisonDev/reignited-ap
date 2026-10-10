# spyrosim — the Spyro Simulator

Plays a Spyro Reignited Trilogy slot against an Archipelago server without the
game. The simulated player clears every level it can reach, sends the same
location ids checkd would send, uses the items the server hands it, and waits
when it can't get further, printing exactly what is in the way.

## Responsible AI Usage Disclosure

This project was written while leveraging open-weight large
language models, primarily self-hosted with open source tools.
All usage happened within a controlled development workflow and
was narrowly spec'd by the developer. This workflow has been
provided including all tools and processes. It can be found in each
project in the [PROCESS.md](https://github.com/THarrisonDev/reignited-ap/blob/main/PROCESS.md) on root.

Remember, these things are tools meant to EMPOWER you never stop thinking.

 **Brains > Bots <3**
 -Tim

Contributors are held to the same expectation. AI-assisted contributions are
welcome, but the human author owns the result: correctness, legibility,
tests, and the decision to merge. Say so in your pull request — disclosure is
there to keep review honest, not to reject a change. We will not be accepting
any code generated in a blind agentic manner and will automatically reject any
PRs directly attributed to AI tools.

## Run

```sh
cd spyrosim
go build
./spyrosim -server localhost:38281 -slot SRTPlayer
```

| Flag | Default | |
| --- | --- | --- |
| `-server` | `localhost:38281` | `host:port`, or a full `ws://`/`wss://` URL (bare hosts try `wss://` then `ws://`) |
| `-slot` | — | slot name (required) |
| `-password` | | room password |
| `-delay` | `1s` | time spent playing each level |
| `-exit` | off | exit when stuck instead of waiting: status 0 if the goal is done, 1 if not |
| `-root` | `.` | the reignited-ap checkout (searched upward) |

On a solo seed nothing else will ever send items, so `-exit` turns a run into
a pass/fail beatability check. Rerunning against the same room resumes: checked
locations come back from the server, and levels are replayed to rebuild the
gem counters.

```
20:24:59 can't get further (Spyro 1: 6 dragons, 0 eggs, 500 gems | Spyro 2: 1 talismans, 3 orbs, 300 gems); waiting for items:
20:24:59   Spyro 1: Peace Keepers: needs 10 dragons (has 6)
20:24:59   Spyro 2: Summer Forest: Swim Ability needs 500 gems for Money Bags (has 300)
20:24:59   Spyro 2: Autumn Plains: needs Spyro 2: Crush's Dungeon: Crush Defeated
```

## Is this config beatable?

```sh
./beatable.sh ~/Rando/ap8/SRTPlayer.yaml 5
```

Generates 5 solo seeds from the yaml with the working-tree world, hosts each
one on a local ArchipelagoServer, and has the sim play it with `-exit`. The
Archipelago AppImage (`$AP_APPIMAGE`, default
`~/Applications/ArchipelagoGenerate`) runs against a private user dir, so the
installed worlds are never touched. A stuck seed prints its stall report and
keeps its logs and spoiler.

## Data

- `client/checkd/locations.json` is the check catalogue (ids, names, types,
  gem thresholds, region order), the same file checkd embeds. A server
  location missing from it is a hard error, which catches drift between the
  apworld and the client.
- `world/data/spyro1.json` gives the homeworlds and balloonist gates;
  `spyro2.json` and `spyro3.json` give the homeworlds and all gate data.
- Items are named through the server's data package, and options, level keys,
  balloonist costs and goals come from slot data.

## What the player obeys

All games: a level opens when its homeworld is open and its key has arrived
(`level_keys`), and homeworlds chain in play order.

- **Spyro 1**:
  - Balloonists read the game's counters: dragons, gems and caught eggs. The
    dragon counter follows received Dragon items under dragonsanity. Each
    balloonist asks for the higher of its vanilla number and the slot's
    `gate_requirements` (shuffled costs), as the mod raises it.
  - Boss portals match `world/spyro1/Game.py`: Toasty needs one Artisans realm
    complete, Twilight Harbor needs Gnorc Cove plus Gnorc Gnexus Delbin, and
    Gnasty Gnorc needs Twilight Harbor. Gnasty's Loot needs Gnasty, 80
    dragons, 12 eggs and 12000 gems.
  - Completion locations are sent when checkd would send them: a level's
    Complete when its exit is reached, a boss's Defeated when it dies, a
    flight's Complete when all 4 target sets are cleared (playing the level
    does all three), and Gnasty's Loot Complete only once the Loot's gems are
    all collected. Portals read those completions from the save, never from
    items, so a flight's Complete check (a real check holding a shuffled item
    when `spyro1_enable_flight_checks` is on) gives whatever the server says.
- **Spyro 2 and 3** play by the gate data in `world/data/spyro2.json` and
  `spyro3.json`, the same fields the logic reads (`world/gates.py` documents
  them; `docs/game_gates.md` lists the gates and their sources). That data
  covers boss-chained homeworlds, talisman/orb/egg doors, the professor door,
  the Sparx realm chain, abilities (received under sanity, bought under
  vanilla), and every Money Bags payment. A payment needs its region open and
  the running total of prices up to it in collected gems. Gem tiers and the
  checks behind paid doors are covered too. Spyro 3 boss kills count as
  eggs.

Not modelled: tricks and skips, and a gate the data doesn't list yet.

## Test

```sh
go test ./...
```

The tests play the real data offline (no server).
