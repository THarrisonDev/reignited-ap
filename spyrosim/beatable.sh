#!/usr/bin/env bash
# Author: Tim Harrison
# Created: 2026-09-26
# Updated: 2026-10-11

# beatable.sh PLAYER.yaml [SEEDS] — generate SEEDS solo seeds from the yaml
# with the working-tree world, host each on a local MultiServer and let
# spyrosim play it to the end. A seed passes when the player reaches the goal.
# The Archipelago AppImage runs against a private user dir, so the installed
# worlds are never touched. Failing seeds keep their logs and spoiler.
set -euo pipefail

yaml=$(realpath "$1")
seeds=${2:-5}
appimage=${AP_APPIMAGE:-$HOME/Applications/ArchipelagoGenerate}
here=$(cd "$(dirname "$0")" && pwd)
root=$(dirname "$here")
work=$(mktemp -d -t spyrosim-XXXX)
server=
# The AppImage runs the server as a grandchild: start it in its own session and
# stop the whole process group, or servers outlive the script.
stop_server() {
	[ -n "$server" ] || return 0
	kill -- "-$server" 2>/dev/null || true
	wait "$server" 2>/dev/null || true
	server=
}
trap stop_server EXIT

mkdir -p "$work/Archipelago/worlds" "$work/players"
(cd "$root/world" && python3 scripts/package_apworld.py \
	"$work/Archipelago/worlds/reignited.apworld" reignited >/dev/null)
cp "$yaml" "$work/players/"
slot=$(sed -n 's/^name: *//p' "$yaml")
(cd "$here" && go build -o "$work/spyrosim" .)

ap() { XDG_DATA_HOME=$work "$appimage" "$@" </dev/null; }

passed=0
for i in $(seq "$seeds"); do
	dir=$work/seed$i
	mkdir -p "$dir"
	if ! ap Generate -- --player_files_path "$work/players" --outputpath "$dir" >"$dir/generate.log" 2>&1; then
		echo "seed $i: generation failed, see $dir/generate.log"
		continue
	fi
	zip=$(ls "$dir"/*.zip)
	(cd "$dir" && unzip -q "$zip" '*_Spoiler.txt')
	port=$((40000 + RANDOM % 20000))
	XDG_DATA_HOME=$work setsid "$appimage" MultiServer -- --host 127.0.0.1 --port "$port" "$zip" \
		</dev/null >"$dir/server.log" 2>&1 &
	server=$!
	for _ in $(seq 120); do
		grep -q "server listening" "$dir/server.log" && break
		sleep 0.5
	done
	if "$work/spyrosim" -root "$root" -server "127.0.0.1:$port" -slot "$slot" -delay 0 -exit >"$dir/run.log" 2>&1; then
		passed=$((passed + 1))
		echo "seed $i: goal reached"
		rm -rf "$dir"
	else
		echo "seed $i: stuck, see $dir/run.log"
		sed -n '/can.t get further/,$p' "$dir/run.log" | head -20
	fi
	stop_server
done
echo "$passed of $seeds seeds beatable ($work)"
[ "$passed" -eq "$seeds" ]
