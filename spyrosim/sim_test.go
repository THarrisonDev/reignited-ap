// Author: Tim Harrison
// Created: 2026-09-26
// Updated: 2026-10-11

package main

import (
	"slices"
	"strings"
	"testing"
)

func testGames(t *testing.T) map[string]*Game {
	t.Helper()
	root, err := findRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	games, err := loadGames(root)
	if err != nil {
		t.Fatal(err)
	}
	return games
}

// newTestSim builds a slot holding every location of its included games.
func newTestSim(t *testing.T, slot SlotData) *Sim {
	t.Helper()
	games := testGames(t)
	var ids []int64
	for _, key := range slot.IncludedGames {
		for _, l := range games[key].Levels {
			for _, c := range l.Checks {
				ids = append(ids, c.ID)
			}
		}
	}
	s, err := NewSim(games, slot, ids)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func playAll(s *Sim) {
	for l := s.Next(); l != nil; l = s.Next() {
		s.Play(l)
	}
}

func slotOf(key string, opts map[string]int, keys map[string]string, goal int64) SlotData {
	return SlotData{
		IncludedGames: []string{key},
		RequiredGames: []string{key},
		Games:         map[string]GameSlot{key: {Options: opts, LevelKeys: keys, GoalLocation: goal}},
	}
}

func reportHas(t *testing.T, s *Sim, want string) {
	t.Helper()
	report := s.Report()
	if !slices.Contains(report, want) {
		t.Fatalf("report lacks %q:\n%s", want, strings.Join(report, "\n"))
	}
}

func TestLevelsSitUnderTheirHomeworld(t *testing.T) {
	games := testGames(t)
	want := map[string]map[string]string{
		"spyro1": {"Artisans": "", "Peace Keepers": "Artisans", "Toasty": "Artisans", "Gnasty's Loot": "Gnorc Gnexus"},
		"spyro2": {"Crush's Dungeon": "Summer Forest", "Winter Tundra": "Autumn Plains", "Canyon Speedway": "Winter Tundra"},
		"spyro3": {"Buzz's Dungeon": "Sunrise Spring", "Super Bonus Round": "Midnight Mountain"},
	}
	for key, levels := range want {
		for _, l := range games[key].Levels {
			parent, ok := levels[l.Name]
			if !ok {
				continue
			}
			got := ""
			if l.Parent != nil {
				got = l.Parent.Name
			}
			if got != parent {
				t.Errorf("%s %s: parent %q, want %q", key, l.Name, got, parent)
			}
		}
	}
}

func TestSpyro1VanillaReachesEverything(t *testing.T) {
	s := newTestSim(t, slotOf("spyro1", map[string]int{"goal": 0}, nil, 103034))
	playAll(s)
	if !s.GoalMet() {
		t.Fatalf("goal not met: %s", s.Progress())
	}
	if got := s.Progress(); got != "Spyro 1: 80 dragons, 12 eggs, 14000 gems" {
		t.Errorf("progress %q", got)
	}
	if report := s.Report(); len(report) != 0 {
		t.Errorf("left over:\n%s", strings.Join(report, "\n"))
	}
}

func TestSpyro1WaitsForLevelKeys(t *testing.T) {
	keys := map[string]string{}
	for _, level := range []string{"Stone Hill", "Dark Hollow", "Town Square", "Sunny Flight", "Toasty"} {
		keys[level] = "Spyro 1: " + level + " Key"
	}
	s := newTestSim(t, slotOf("spyro1", map[string]int{"level_locks": 1}, keys, 0))
	playAll(s)
	reportHas(t, s, "Spyro 1: Stone Hill: needs Spyro 1: Stone Hill Key")
	reportHas(t, s, "Spyro 1: Peace Keepers: needs 10 dragons (has 4)")

	s.Receive("Spyro 1: Stone Hill Key")
	playAll(s)
	reportHas(t, s, "Spyro 1: Toasty: needs Spyro 1: Toasty Key")
	reportHas(t, s, "Spyro 1: Peace Keepers: needs 10 dragons (has 8)")
	if !s.Got("Spyro 1: Stone Hill Complete") {
		t.Error("Stone Hill not played after its key arrived")
	}
}

func TestSpyro1DragonsanityCountsDragonItems(t *testing.T) {
	s := newTestSim(t, slotOf("spyro1", map[string]int{"dragon_shuffle": 1}, nil, 0))
	playAll(s)
	reportHas(t, s, "Spyro 1: Peace Keepers: needs 10 dragons (has 0)")
	for range 10 {
		s.Receive("Spyro 1: Dragon")
	}
	if s.Next() == nil {
		t.Fatal("10 Dragon items did not open Peace Keepers")
	}
}

func TestSpyro1ShuffledBalloonistsRaiseOnly(t *testing.T) {
	slot := slotOf("spyro1", map[string]int{"dragon_shuffle": 1, "balloonist_gates": 1}, nil, 0)
	g := slot.Games["spyro1"]
	g.GateRequirements = map[string]Gate{
		"Peace Keepers":  {Type: "dragons", Needed: 11},
		"Magic Crafters": {Type: "gems", Needed: 900},
	}
	slot.Games["spyro1"] = g
	s := newTestSim(t, slot)
	for range 10 {
		s.Receive("Spyro 1: Dragon")
	}
	playAll(s)
	reportHas(t, s, "Spyro 1: Peace Keepers: needs 11 dragons (has 10)")
	s.Receive("Spyro 1: Dragon")
	if s.Next() == nil {
		t.Fatal("11 Dragon items did not open Peace Keepers")
	}
	for _, l := range s.games[0].Levels {
		if l.Name == "Magic Crafters" {
			if why := s.Shut(l); !strings.Contains(why, "1200 gems") {
				t.Errorf("a gate below vanilla changed Magic Crafters: %q", why)
			}
		}
	}
}

func TestSpyro1ClearingArtisansCollectsItsChecks(t *testing.T) {
	s := newTestSim(t, slotOf("spyro1", nil, nil, 0))
	s.Play(s.Next())
	pending := s.TakePending()
	for _, id := range []int64{100000, 102013} {
		if !slices.Contains(pending, id) {
			t.Errorf("clearing Artisans did not collect %d: %v", id, pending)
		}
	}
}

func TestOnlySlotLocationsAreSent(t *testing.T) {
	games := testGames(t)
	s, err := NewSim(games, slotOf("spyro1", nil, nil, 0), []int64{100000})
	if err != nil {
		t.Fatal(err)
	}
	s.Play(s.Next())
	if got := s.TakePending(); !slices.Equal(got, []int64{100000}) {
		t.Errorf("sent %v, want only the slot's 100000", got)
	}
	if got := s.TakePending(); len(got) != 0 {
		t.Errorf("sent %v twice", got)
	}
}

func TestSpyro2VanillaBuysItsAbilities(t *testing.T) {
	s := newTestSim(t, slotOf("spyro2", map[string]int{"goal": 1}, nil, 226000))
	playAll(s)
	if !s.GoalMet() {
		t.Fatalf("Ripto not reached: %s\n%s", s.Progress(), strings.Join(s.Report(), "\n"))
	}
	if report := s.Report(); len(report) != 0 {
		t.Errorf("left over:\n%s", strings.Join(report, "\n"))
	}
}

func TestSpyro2SanityWaitsForAbilityItems(t *testing.T) {
	s := newTestSim(t, slotOf("spyro2", map[string]int{"goal": 1, "moneybags_abilities": 1}, nil, 226000))
	playAll(s)
	reportHas(t, s, "Spyro 2: Hurricos: needs Spyro 2: Swim Ability")
	if !s.Got("Spyro 2: Summer Forest: Swim Ability") {
		t.Error("Swim purchase not made: the location has no item rule")
	}
	for _, a := range []string{"Swim", "Climb", "Headbash"} {
		s.Receive("Spyro 2: " + a + " Ability")
	}
	playAll(s)
	if !s.GoalMet() {
		t.Fatalf("Ripto not reached with every ability:\n%s", strings.Join(s.Report(), "\n"))
	}
}

func TestSpyro2GemTiers(t *testing.T) {
	s := newTestSim(t, slotOf("spyro2", map[string]int{"moneybags_abilities": 1}, nil, 0))
	playAll(s)
	reportHas(t, s, "Spyro 2: Glimmer: Gem lamp flight in cave, All Gems needs Spyro 2: Climb Ability")
}

func TestSpyro2MoneyBagsWantsGems(t *testing.T) {
	keys := map[string]string{}
	for _, l := range testGames(t)["spyro2"].Levels {
		if !l.IsHub {
			keys[l.Name] = "Spyro 2: " + l.Name + " Key"
		}
	}
	s := newTestSim(t, slotOf("spyro2", map[string]int{"level_locks": 1}, keys, 0))
	s.Receive("Spyro 2: Glimmer Key")
	playAll(s)
	reportHas(t, s, "Spyro 2: Summer Forest: Swim Ability needs 600 gems to pay Money Bags for Swim (has 300)")
	s.Receive("Spyro 2: Colossus Key")
	playAll(s)
	if !s.Got("Spyro 2: Summer Forest: Swim Ability") {
		t.Errorf("Swim not bought with %s", s.Progress())
	}
}

func TestSpyro3ReachesTheSorceress(t *testing.T) {
	s := newTestSim(t, slotOf("spyro3", map[string]int{"goal": 1}, nil, 334000))
	playAll(s)
	if !s.GoalMet() {
		t.Fatalf("Sorceress not reached: %s\n%s", s.Progress(), strings.Join(s.Report(), "\n"))
	}
	if got := s.Progress(); !strings.HasPrefix(got, "Spyro 3: 150 eggs") {
		t.Errorf("progress %q", got)
	}
}

func lockedSlot(t *testing.T, key string) SlotData {
	t.Helper()
	keys := map[string]string{}
	for _, l := range testGames(t)[key].Levels {
		if !l.IsHub {
			keys[l.Name] = l.Game.Prefix + l.Name + " Key"
		}
	}
	return slotOf(key, map[string]int{"level_locks": 1}, keys, 0)
}

func TestSpyro3EggPortalsAndCompanions(t *testing.T) {
	s := newTestSim(t, lockedSlot(t, "spyro3"))
	s.Receive("Spyro 3: Molten Crater Key")
	s.Receive("Spyro 3: Sheila's Alp Key")
	playAll(s)
	reportHas(t, s, "Spyro 3: Molten Crater: needs 10 eggs (has 8)")
	if !s.Got("Spyro 3: Sheila's Alp: Nan") {
		t.Error("Sheila's Alp not played: Sunrise Spring's 400 gems pay Sheila's 300")
	}
	s.Receive("Spyro 3: Sunny Villa Key")
	playAll(s)
	if !s.Got("Spyro 3: Molten Crater: Curlie") {
		t.Errorf("Molten Crater still shut with %s", s.Progress())
	}
}

func TestSpyro3EggPortalsCountEggItems(t *testing.T) {
	s := newTestSim(t, slotOf("spyro3", map[string]int{"egg_shuffle": 1}, nil, 0))
	playAll(s)
	// With egg_shuffle on, Buzz's Dungeon gates on its hub's five sibling
	// levels completing, which themselves need eggs. With zero items received,
	// the count is genuinely 0, not a free 3 from unshuffled boss dungeons.
	reportHas(t, s, "Spyro 3: Molten Crater: needs 10 eggs (has 0)")
	for range 10 {
		s.Receive("Spyro 3: Dragon Egg")
	}
	if s.Next() == nil {
		t.Fatal("10 Dragon Egg items did not open Molten Crater")
	}
}

func TestSpyro2OrbDoors(t *testing.T) {
	s := newTestSim(t, lockedSlot(t, "spyro2"))
	s.Receive("Spyro 2: Glimmer Key")
	s.Receive("Spyro 2: Colossus Key")
	s.Receive("Spyro 2: Ocean Speedway Key")
	playAll(s)
	if !s.Got("Spyro 2: Ocean Speedway: Follow Hunter") {
		t.Errorf("Ocean Speedway shut with Swim and %s:\n%s", s.Progress(), strings.Join(s.Report(), "\n"))
	}
	s2 := newTestSim(t, lockedSlot(t, "spyro2"))
	s2.Receive("Spyro 2: Glimmer Key")
	s2.Receive("Spyro 2: Ocean Speedway Key")
	playAll(s2)
	reportHas(t, s2, "Spyro 2: Ocean Speedway: needs Spyro 2: Swim Ability")
}

// playServed plays like playAll, with the server handing out the item each
// sent location holds.
func playServed(s *Sim, holds map[int64]string) {
	for l := s.Next(); l != nil; l = s.Next() {
		s.Play(l)
		for _, id := range s.TakePending() {
			if item, ok := holds[id]; ok {
				s.Receive(item)
			}
		}
	}
}

func TestSpyro1FlightCompleteHoldsAnItem(t *testing.T) {
	keys := map[string]string{}
	for _, level := range []string{"Stone Hill", "Dark Hollow", "Town Square", "Sunny Flight", "Toasty"} {
		keys[level] = "Spyro 1: " + level + " Key"
	}
	holds := map[int64]string{
		103006: "Spyro 1: Stone Hill Key", // Sunny Flight Complete.
		100011: "Spyro 1: Dark Hollow Key",
		100012: "Spyro 1: Town Square Key",
		100013: "Spyro 1: Toasty Key",
	}
	s := newTestSim(t, slotOf("spyro1", map[string]int{"level_locks": 1}, keys, 103034))
	playServed(s, holds)
	reportHas(t, s, "Spyro 1: Sunny Flight: needs Spyro 1: Sunny Flight Key")
	if s.Has("Spyro 1: Sunny Flight Complete") {
		t.Error("the flight's own event item was granted")
	}

	s.Receive("Spyro 1: Sunny Flight Key")
	playServed(s, holds)
	if !s.GoalMet() {
		t.Fatalf("goal not met with the flight check holding the Stone Hill Key: %s\n%s",
			s.Progress(), strings.Join(s.Report(), "\n"))
	}
	if s.Has("Spyro 1: Sunny Flight Complete") {
		t.Error("the flight's own event item was granted")
	}
}

func TestSpyro1LootGoalNeedsTheLootGems(t *testing.T) {
	s := newTestSim(t, slotOf("spyro1", map[string]int{"goal": 1}, nil, 103035))
	var loot *Level
	for _, l := range s.games[0].Levels {
		if l.Name == "Gnasty's Loot" {
			loot = l
		}
	}
	var complete Check
	for _, c := range loot.Checks {
		if c.Location == "Spyro 1: Gnasty's Loot Complete" {
			complete = c
		}
	}
	s.gems[loot] = 75
	if why := s.game("spyro1").rules.need(s, loot, complete); why != "needs 100 percent of Gnasty's Loot gems (has 75)" {
		t.Errorf("Loot Complete at 75%% gems: %q", why)
	}
	s.gems[loot] = 0
	playAll(s)
	if !s.GoalMet() || s.gems[loot] != 100 {
		t.Fatalf("Loot goal not met (%d%% Loot gems): %s\n%s", s.gems[loot], s.Progress(), strings.Join(s.Report(), "\n"))
	}
}
