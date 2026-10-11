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

.PHONY: bundle bundle-world bundle-client bundle-poptracker clean

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

clean:
	rm -rf $(DIST)
