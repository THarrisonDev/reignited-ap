# Reignited AP — release bundle assembly.
#
# Builds the single player-facing download: the client (launcher, save scanner,
# UE4SS mod), the world .apworld, the pinned UE4SS fallback, and START-HERE.
# See world/RELEASE.md for the release order and the checks to run first.

VERSION      := $(shell sed -n 's/.*"world_version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' world/archipelago.json)
WORLD_NAME   := reignited
PACKTRACKER  := spyro_reignited_ap
UE4SS_VER    := v3.0.1
UE4SS_URL    := https://github.com/UE4SS-RE/RE-UE4SS/releases/download/$(UE4SS_VER)/UE4SS_$(UE4SS_VER).zip
UE4SS_ZIP    ?= bundle/UE4SS_$(UE4SS_VER).zip

DIST         := dist
BUNDLE       := $(DIST)/$(WORLD_NAME)-ap_$(VERSION)
BUNDLE_ZIP   := $(BUNDLE)_bundle.zip
# Windows-only variant: the same bundle with the native Linux launcher
# stripped, matching the published `_bundle-win.zip`/`.7z` artifacts.
BUNDLE_WIN_ZIP := $(BUNDLE)_bundle-win.zip
BUNDLE_WIN_7Z  := $(BUNDLE)_bundle-win.7z

# Copy-paste detector (node); `make duplication` runs it over the whole tree.
JSCPD        ?= npx --yes jscpd

.PHONY: bundle bundle-world bundle-client bundle-poptracker bundle-win quality duplication clean

# Assemble the whole bundle. Override UE4SS_ZIP to use a local copy instead of
# downloading the pinned release.
bundle: bundle-world bundle-client bundle-poptracker $(UE4SS_ZIP)
	@if [ -z "$(VERSION)" ]; then echo "ERROR: no world_version in world/archipelago.json" >&2; exit 1; fi
	rm -rf $(BUNDLE)
	mkdir -p $(BUNDLE)/world $(BUNDLE)/UE4SS $(BUNDLE)/poptracker
	cp -f world/$(WORLD_NAME)_$(VERSION).apworld $(BUNDLE)/world/
	cp -a client/dist/client $(BUNDLE)/client
	cp -a poptracker-pack/dist/$(PACKTRACKER) $(BUNDLE)/poptracker/
	cp -f bundle/POPTRACKER-README.txt $(BUNDLE)/poptracker/README.txt
	cp -f $(UE4SS_ZIP) $(BUNDLE)/UE4SS/
	sed 's/@VERSION@/$(VERSION)/g' bundle/START-HERE.txt > $(BUNDLE)/START-HERE.txt
	cd $(DIST) && rm -f $(notdir $(BUNDLE_ZIP)) && zip -q -r $(notdir $(BUNDLE_ZIP)) $(notdir $(BUNDLE))
	@echo "packaged $(abspath $(BUNDLE_ZIP))"

# Windows-only artifact: stage a copy of the bundle, drop the native Linux
# launcher, and archive it under the same top-level release name so it matches
# the full bundle exactly apart from `client/srtap_launcher`.
bundle-win: bundle
	rm -rf $(DIST)/.bundle-win
	mkdir -p $(DIST)/.bundle-win
	cp -a $(BUNDLE) $(DIST)/.bundle-win/
	rm -f $(DIST)/.bundle-win/$(notdir $(BUNDLE))/client/srtap_launcher
	cd $(DIST)/.bundle-win && rm -f ../$(notdir $(BUNDLE_WIN_ZIP)) \
		&& zip -q -r ../$(notdir $(BUNDLE_WIN_ZIP)) $(notdir $(BUNDLE))
	cd $(DIST)/.bundle-win && rm -f ../$(notdir $(BUNDLE_WIN_7Z)) \
		&& 7z a -y -bd -bso0 -bsp0 ../$(notdir $(BUNDLE_WIN_7Z)) $(notdir $(BUNDLE)) >/dev/null
	rm -rf $(DIST)/.bundle-win
	@echo "packaged $(abspath $(BUNDLE_WIN_ZIP)) and $(abspath $(BUNDLE_WIN_7Z))"

# The versioned .apworld lives beside the world package (both gitignored).
bundle-world:
	$(MAKE) --no-print-directory -C world package
	cp -f world/$(WORLD_NAME).apworld world/$(WORLD_NAME)_$(VERSION).apworld

bundle-client:
	$(MAKE) --no-print-directory -C client package-client

bundle-poptracker:
	$(MAKE) --no-print-directory -C poptracker-pack package

$(UE4SS_ZIP):
	mkdir -p $(dir $@)
	curl -fL -o $@ $(UE4SS_URL)

# Whole-repo quality sweep: every linter plus the dead-code and duplication
# scans that no single-language linter covers. Each submodule runs its own
# toolchain, so a finding is reported against the repo that owns the fix:
#   world   -- ruff (correctness/bugs/complexity) + vulture (unreferenced names)
#   client  -- golangci-lint (unused/unparam/cyclop/...) + luacheck (unused
#              locals, complexity) + the Lua module audit + the generated-file
#              audit
#   spyrosim -- golangci-lint over the simulator
#   poptracker -- generated files match the world data
# also: the AP minimum version the docs advertise matches archipelago.json
# plus jscpd: zero clones at >=5 lines / >=50 tokens across all four repos.
# Install once: pip install ruff vulture; golangci-lint; luarocks install
# luacheck; node (for npx jscpd).
quality:
	python3 scripts/audit_ap_version.py
	$(MAKE) --no-print-directory -C world check-version lint deadcode
	$(MAKE) --no-print-directory -C client lint deadcode audit
	$(MAKE) --no-print-directory -C poptracker-pack check
	$(MAKE) --no-print-directory duplication
	cd spyrosim && if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found; running go vet only" >&2; \
		go vet ./...; \
	fi

# Copy-paste check: jscpd with the repo config (.jscpd.json) fails on any
# clone. Generated data (regions.lua, locations.json, the pack maps) is
# excluded there, since it is emitted by the generators, not written.
duplication:
	$(JSCPD) world client spyrosim poptracker-pack --config .jscpd.json

# Remove every build product, including the root-level bundle zips that the
# release targets drop here (gitignored, and nothing else deletes them).
# bundle/UE4SS_*.zip is a cached download (a build input), so it stays.
clean:
	rm -rf $(DIST)
	rm -f reignited-ap_*_bundle.zip reignited-ap_*_bundle-win.zip reignited-ap_*_bundle-win.7z
