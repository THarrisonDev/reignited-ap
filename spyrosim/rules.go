// Author: Tim Harrison
// Created: 2026-09-26
// Updated: 2026-10-11

package main

import (
	"fmt"
	"strconv"
)

// needDone wants every named location of the game collected.
func needDone(s *Sim, g *Game, locations ...string) string {
	for _, name := range locations {
		if !s.Got(g.Prefix + name) {
			return "needs " + g.Prefix + name
		}
	}
	return ""
}

func atLeast(what string, have, need int) string {
	if have >= need {
		return ""
	}
	return fmt.Sprintf("needs %d %s (has %d)", need, what, have)
}

// Spyro 1: balloonists read the game's counters; boss portals follow
// world/spyro1/Game.py. The mod raises each balloonist to the slot's
// gate_requirements (never lowers one), so the higher number wins.
type spyro1 struct{ gates map[string]Gate }

func (r spyro1) gate(s *Sim, l *Level) string {
	g := l.Game
	if gate, ok := r.gates[l.Name]; ok {
		return r.balloonist(s, g, l.Name, gate)
	}
	switch l.Name {
	case "Toasty":
		return needAnyDone(s, g, "needs one Artisans realm complete",
			"Stone Hill Complete", "Dark Hollow Complete", "Town Square Complete")
	case "Twilight Harbor":
		return needDone(s, g, "Gnorc Cove Complete", "Gnorc Gnexus: Free Delbin")
	case "Gnasty Gnorc":
		return needDone(s, g, "Twilight Harbor Complete")
	case "Gnasty's Loot":
		return lootGate(s, g)
	}
	return ""
}

// balloonist is the balloonist gate of one level: the slot's raised
// requirement when it is higher and of the same type, read off the game's
// counter.
func (r spyro1) balloonist(s *Sim, g *Game, name string, gate Gate) string {
	if req, ok := s.opts(g).GateRequirements[name]; ok && req.Type == gate.Type && req.Needed > gate.Needed {
		gate.Needed = req.Needed
	}
	return atLeast(gate.Type, s.counter(g, gate.Type), gate.Needed)
}

// counter is the game's counter a gate names: gems add up the levels' gems,
// every other type counts collected checks.
func (s *Sim) counter(g *Game, kind string) int {
	if kind == "gems" {
		return s.gemCount(g)
	}
	return s.count(g, kind)
}

// needAnyDone wants at least one of the named locations collected.
func needAnyDone(s *Sim, g *Game, why string, locations ...string) string {
	for _, name := range locations {
		if s.Got(g.Prefix + name) {
			return ""
		}
	}
	return why
}

// lootGate is Gnasty's Loot's gate: the boss beaten, then the counters the
// loot door shows.
func lootGate(s *Sim, g *Game) string {
	if why := needDone(s, g, "Gnasty Gnorc Defeated"); why != "" {
		return why
	}
	if why := atLeast("dragons", s.count(g, "dragons"), 80); why != "" {
		return why
	}
	if why := atLeast("eggs", s.count(g, "eggs"), 12); why != "" {
		return why
	}
	return atLeast("gems", s.gemCount(g), 12000)
}

// need mirrors when checkd sends a Spyro 1 event location: a level's Complete
// once its exit vortex is reached, a boss's Defeated once it dies, a flight's
// Complete once all 4 target sets are cleared (a real check holding a shuffled
// item, not an event), all of which playing the level does; Gnasty's Loot
// Complete only once the Loot's gems are all collected.
func (spyro1) need(s *Sim, l *Level, c Check) string {
	if c.Type == KindEvent && l.Type == "BONUS" {
		return atLeast("percent of "+l.Name+" gems", s.gems[l], 100)
	}
	return ""
}

// GateData is a tabular game's gate section of world/data/<game>.json; see
// world/gates.py for the meaning, which tabular mirrors.
type GateData struct {
	Homeworlds []string                      `json:"homeworlds"`
	Payments   []Payment                     `json:"payments"`
	Gates      map[string]GateReq            `json:"gates"`
	CheckGates map[string]GateReq            `json:"checkGates"`
	GemGates   map[string]map[string]GateReq `json:"gemGates"`
}

type Payment struct {
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Region   string `json:"region"`
	Location string `json:"location"`
}

type GateReq struct {
	Bosses    []string `json:"bosses"`
	Abilities []string `json:"abilities"`
	Levels    []string `json:"levels"`
	Payments  []string `json:"payments"`
	Talismans int      `json:"talismans"`
	Orbs      int      `json:"orbs"`
	Eggs      int      `json:"eggs"`
	Gems      int      `json:"gems"`
}

// tabular runs Spyro 2 and 3 from their gate data. Bosses are the boss checks
// collected, levels are open portals, counts are collected checks (Spyro 3
// boss kills are eggs) and a payment needs its region open plus the running
// total of prices up to it in collected gems.
type tabular struct {
	GateData
	runningTotal map[string]int
	payment      map[string]Payment
}

func newTabular(d GateData) tabular {
	t := tabular{GateData: d, runningTotal: map[string]int{}, payment: map[string]Payment{}}
	total := 0
	for _, p := range d.Payments {
		total += p.Price
		t.runningTotal[p.Name] = total
		t.payment[p.Name] = p
	}
	return t
}

func (t tabular) level(g *Game, name string) *Level {
	for _, l := range g.Levels {
		if l.Name == name {
			return l
		}
	}
	return nil
}

// ability is owned once received (sanity) or once bought from Money Bags.
func (t tabular) ability(s *Sim, g *Game, name string) bool {
	if s.opts(g).Options["moneybags_abilities"] == 1 {
		return s.Has(g.Prefix + name + " Ability")
	}
	for _, p := range t.Payments {
		if p.Name == name && p.Location != "" {
			return s.Got(g.Prefix + p.Location)
		}
	}
	return false
}

func (t tabular) pay(s *Sim, g *Game, name string) string {
	p := t.payment[name]
	if why := s.Shut(t.level(g, p.Region)); why != "" {
		return fmt.Sprintf("needs %s open to pay Money Bags (%s)", p.Region, name)
	}
	return atLeast("gems to pay Money Bags for "+name, s.gemCount(g), t.runningTotal[name])
}

func (t tabular) met(s *Sim, g *Game, req GateReq) string {
	if why := t.metBosses(s, g, req.Bosses); why != "" {
		return why
	}
	if why := t.metAbilities(s, g, req.Abilities); why != "" {
		return why
	}
	if why := t.metLevels(s, g, req.Levels); why != "" {
		return why
	}
	if why := t.metPayments(s, g, req.Payments); why != "" {
		return why
	}
	return t.metCounts(s, g, req)
}

// metBosses wants every named boss defeated.
func (t tabular) metBosses(s *Sim, g *Game, bosses []string) string {
	for _, boss := range bosses {
		if !s.Got(t.bossLocation(g, boss)) {
			return "needs " + g.Prefix + boss
		}
	}
	return ""
}

// metAbilities wants every named ability owned.
func (t tabular) metAbilities(s *Sim, g *Game, abilities []string) string {
	for _, a := range abilities {
		if !t.ability(s, g, a) {
			return "needs " + g.Prefix + a + " Ability"
		}
	}
	return ""
}

// metLevels wants every named level open.
func (t tabular) metLevels(s *Sim, g *Game, levels []string) string {
	for _, name := range levels {
		if s.Shut(t.level(g, name)) != "" {
			return "needs " + g.Prefix + name + " open"
		}
	}
	return ""
}

// metPayments wants every named Money Bags payment affordable.
func (t tabular) metPayments(s *Sim, g *Game, payments []string) string {
	for _, name := range payments {
		if why := t.pay(s, g, name); why != "" {
			return why
		}
	}
	return ""
}

// metCounts wants the counters the request names to be high enough.
func (t tabular) metCounts(s *Sim, g *Game, req GateReq) string {
	for _, c := range []struct {
		what       string
		have, need int
	}{
		{"talismans", s.count(g, "talismans"), req.Talismans},
		{"orbs", s.count(g, "orbs"), req.Orbs},
		{"eggs", s.count(g, "eggs"), req.Eggs},
		{"gems", s.gemCount(g), req.Gems},
	} {
		if why := atLeast(c.what, c.have, c.need); why != "" {
			return why
		}
	}
	return ""
}

func (t tabular) bossLocation(g *Game, boss string) string {
	for _, l := range g.Levels {
		for _, c := range l.Checks {
			if c.Type == KindBoss && c.Name == boss {
				return c.Location
			}
		}
	}
	return ""
}

func (t tabular) gate(s *Sim, l *Level) string {
	return t.met(s, l.Game, t.Gates[l.Name])
}

func (t tabular) need(s *Sim, l *Level, c Check) string {
	g := l.Game
	name := c.Location[len(g.Prefix):]
	if c.Type == KindGem {
		for _, tier := range []int{25, 50, 75, 100} {
			if tier > c.Threshold {
				break
			}
			if why := t.met(s, g, t.GemGates[l.Name][strconv.Itoa(tier)]); why != "" {
				return why
			}
		}
		return ""
	}
	for _, p := range t.Payments {
		if p.Location == name {
			return t.pay(s, g, p.Name)
		}
	}
	return t.met(s, g, t.CheckGates[name])
}
