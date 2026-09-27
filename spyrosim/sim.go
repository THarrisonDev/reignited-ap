package main

import (
	"fmt"
	"strings"
)

type rules interface {
	gate(s *Sim, l *Level) string
	need(s *Sim, l *Level, c Check) string
}

type GameSlot struct {
	Options          map[string]int    `json:"options"`
	LevelKeys        map[string]string `json:"level_keys"`
	GateRequirements map[string]Gate   `json:"gate_requirements"`
	GoalLocation     int64             `json:"goal_location"`
}

type SlotData struct {
	IncludedGames []string            `json:"included_games"`
	RequiredGames []string            `json:"required_games"`
	Games         map[string]GameSlot `json:"games"`
}

// Sim is the simulated player: what it holds, what its save has collected,
// and which of those collections the server has yet to hear about.
type Sim struct {
	slot    SlotData
	games   []*Game
	items   map[string]int
	have    map[string]bool
	gems    map[*Level]int // Percent of the level's gems collected.
	byID    map[int64]string
	inSlot  map[int64]bool
	pending []int64
}

func NewSim(all map[string]*Game, slot SlotData, slotLocations []int64) (*Sim, error) {
	s := &Sim{
		slot:   slot,
		items:  map[string]int{},
		have:   map[string]bool{},
		gems:   map[*Level]int{},
		byID:   map[int64]string{},
		inSlot: map[int64]bool{},
	}
	for _, key := range slot.IncludedGames {
		g, ok := all[key]
		if !ok {
			return nil, fmt.Errorf("slot includes %q, which %s does not describe", key, locationsFile)
		}
		s.games = append(s.games, g)
		for _, l := range g.Levels {
			for _, c := range l.Checks {
				s.byID[c.ID] = c.Location
			}
		}
	}
	for _, id := range slotLocations {
		if _, ok := s.byID[id]; !ok {
			return nil, fmt.Errorf("server location %d is not in %s", id, locationsFile)
		}
		s.inSlot[id] = true
	}
	return s, nil
}

func (s *Sim) Receive(item string) { s.items[item]++ }

func (s *Sim) Has(item string) bool { return s.items[item] > 0 }

func (s *Sim) Got(location string) bool { return s.have[location] }

// MarkChecked records locations the server already has, such as from an
// earlier run.
func (s *Sim) MarkChecked(ids []int64) {
	for _, id := range ids {
		if name, ok := s.byID[id]; ok {
			s.have[name] = true
		}
	}
}

func (s *Sim) opts(g *Game) GameSlot { return s.slot.Games[g.Key] }

// Shut says why the player cannot enter the level, or "" when they can.
func (s *Sim) Shut(l *Level) string {
	if l.Parent != nil {
		if why := s.Shut(l.Parent); why != "" {
			return l.Parent.Name + " is shut"
		}
	}
	if key, locked := s.opts(l.Game).LevelKeys[l.Name]; locked && !s.Has(key) {
		return "needs " + key
	}
	return l.Game.rules.gate(s, l)
}

func (s *Sim) collectible(l *Level, c Check) bool {
	return !s.have[c.Location] && l.Game.rules.need(s, l, c) == ""
}

// gemCap is the percent of the level's gems the player can collect now: the
// highest gem check within reach, or all of them when the level has none.
func (s *Sim) gemCap(l *Level) int {
	limit, capped := 0, false
	for _, c := range l.Checks {
		if c.Type != "gem" {
			continue
		}
		capped = true
		if c.Threshold > limit && l.Game.rules.need(s, l, c) == "" {
			limit = c.Threshold
		}
	}
	if !capped {
		return 100
	}
	return limit
}

func (s *Sim) worthVisiting(l *Level) bool {
	if s.Shut(l) != "" {
		return false
	}
	if l.GemTotal > 0 && s.gems[l] < s.gemCap(l) {
		return true
	}
	for _, c := range l.Checks {
		if s.collectible(l, c) {
			return true
		}
	}
	return false
}

// Next is the first level, in play order, where the player can collect
// something new, or nil when the player can't get further.
func (s *Sim) Next() *Level {
	for _, g := range s.games {
		for _, l := range g.Levels {
			if s.worthVisiting(l) {
				return l
			}
		}
	}
	return nil
}

// Play clears everything the player can reach in the level and returns the
// slot locations the save gained.
func (s *Sim) Play(l *Level) []string {
	if l.GemTotal > 0 {
		s.gems[l] = max(s.gems[l], s.gemCap(l))
	}
	var got []string
	for progress := true; progress; {
		progress = false
		for _, c := range l.Checks {
			if s.collectible(l, c) {
				if s.collect(c.Location, c.ID) {
					got = append(got, c.Location)
				}
				progress = true
			}
		}
	}
	return got
}

// collect saves the location and reports whether the slot owns it to send.
func (s *Sim) collect(location string, id int64) bool {
	s.have[location] = true
	if s.inSlot[id] {
		s.pending = append(s.pending, id)
	}
	return s.inSlot[id]
}

// TakePending returns and clears the collected location ids the server has
// not yet been sent.
func (s *Sim) TakePending() []int64 {
	ids := s.pending
	s.pending = nil
	return ids
}

// count is the game's own counter: collected checks of a type. When the
// matching collectible is shuffled, the counter follows received items instead,
// because the mod reconciles the game's counters to them. Bosses always add to
// the egg count because their kill is the boss dungeon's egg record.
func (s *Sim) count(g *Game, kind string) int {
	if g.Key == "spyro1" && kind == "dragons" && s.opts(g).Options["dragon_shuffle"] == 1 {
		return s.items[g.Prefix+"Dragon"]
	}
	if g.Key == "spyro2" && kind == "orbs" && s.opts(g).Options["orb_shuffle"] == 1 {
		return s.items[g.Prefix+"Orb"]
	}
	eggShuffle := g.Key == "spyro3" && kind == "eggs" && s.opts(g).Options["egg_shuffle"] == 1
	types := map[string][]Kind{
		"dragons":   {KindDragon},
		"eggs":      {KindEgg},
		"orbs":      {KindOrb},
		"talismans": {KindTalisman},
	}[kind]
	if g.Key == "spyro3" && kind == "eggs" {
		types = []Kind{KindBoss}
		if !eggShuffle {
			types = append(types, KindEgg)
		}
	}
	n := 0
	if eggShuffle {
		n = s.items[g.Prefix+"Dragon Egg"]
	}
	for _, l := range g.Levels {
		for _, c := range l.Checks {
			for _, t := range types {
				if c.Type == t && s.have[c.Location] {
					n++
				}
			}
		}
	}
	return n
}

func (s *Sim) gemCount(g *Game) int {
	n := 0
	for _, l := range g.Levels {
		n += l.GemTotal * s.gems[l] / 100
	}
	return n
}

// GoalMet reports whether every required game's goal is done.
func (s *Sim) GoalMet() bool {
	for _, key := range s.slot.RequiredGames {
		gs := s.slot.Games[key]
		if key == "spyro1" && gs.Options["goal"] == 2 {
			g := s.game(key)
			if !s.Got(g.Prefix+"Gnasty Gnorc Defeated") || s.count(g, "dragons") < 80 {
				return false
			}
			continue
		}
		if name, ok := s.byID[gs.GoalLocation]; !ok || !s.have[name] {
			return false
		}
	}
	return len(s.slot.RequiredGames) > 0
}

func (s *Sim) game(key string) *Game {
	for _, g := range s.games {
		if g.Key == key {
			return g
		}
	}
	return nil
}

// Report describes where the player is stuck: shut homeworlds and portals
// that still hold checks, and open levels holding unreachable checks.
func (s *Sim) Report() []string {
	var lines []string
	for _, g := range s.games {
		for _, l := range g.Levels {
			var left []Check
			for _, c := range l.Checks {
				if !s.have[c.Location] && s.inSlot[c.ID] {
					left = append(left, c)
				}
			}
			if why := s.Shut(l); why != "" {
				if (l.Parent == nil || s.Shut(l.Parent) == "") && (l.IsHub || len(left) > 0) {
					lines = append(lines, fmt.Sprintf("%s%s: %s", g.Prefix, l.Name, why))
				}
				continue
			}
			var reasons []string
			byReason := map[string][]string{}
			for _, c := range left {
				why := g.rules.need(s, l, c)
				if byReason[why] == nil {
					reasons = append(reasons, why)
				}
				byReason[why] = append(byReason[why], strings.TrimPrefix(c.Location, g.Prefix+l.Name+": "))
			}
			for _, why := range reasons {
				lines = append(lines, fmt.Sprintf("%s%s: %s %s", g.Prefix, l.Name, strings.Join(byReason[why], ", "), why))
			}
		}
	}
	return lines
}

// Progress summarises the counters the game shows the player.
func (s *Sim) Progress() string {
	var parts []string
	for _, g := range s.games {
		var p string
		switch g.Key {
		case "spyro1":
			p = fmt.Sprintf("%d dragons, %d eggs, %d gems", s.count(g, "dragons"), s.count(g, "eggs"), s.gemCount(g))
		case "spyro2":
			p = fmt.Sprintf("%d talismans, %d orbs, %d gems", s.count(g, "talismans"), s.count(g, "orbs"), s.gemCount(g))
		case "spyro3":
			p = fmt.Sprintf("%d eggs, %d gems", s.count(g, "eggs"), s.gemCount(g))
		}
		parts = append(parts, g.Prefix+p)
	}
	return strings.Join(parts, " | ")
}
