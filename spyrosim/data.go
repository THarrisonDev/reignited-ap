// Author: Tim Harrison
// Created: 2026-09-26
// Updated: 2026-10-11

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	locationsFile = "client/checkd/locations.json"
	dataFile      = "world/data/%s.json"
)

// Kind enumerates the check kinds in checkd's locations.json (the values
// gen_locations.py writes); sim.go's count() maps counter kinds to these.
type Kind string

const (
	KindDragon   Kind = "dragon"
	KindEgg      Kind = "egg"
	KindGem      Kind = "gem"
	KindOrb      Kind = "orb"
	KindTalisman Kind = "talisman"
	KindBoss     Kind = "boss"
	KindEvent    Kind = "event"
)

type Check struct {
	Type      Kind   `json:"type"`
	Name      string `json:"name"`
	Location  string `json:"location"`
	ID        int64  `json:"id"`
	Threshold int    `json:"threshold"`
	Ability   string `json:"ability"`
}

type Level struct {
	Name     string
	Game     *Game
	Type     string // Spyro 1 only: world/data/spyro1.json's HUB|NORMAL|FLIGHT|BOSS|BONUS
	IsHub    bool
	Parent   *Level // the homeworld a level's portal stands in; the previous homeworld for a homeworld
	GemTotal int
	Checks   []Check
}

type Game struct {
	Key    string
	Prefix string
	Levels []*Level
	rules  rules
}

type Gate struct {
	Type   string `json:"type"`
	Needed int    `json:"needed"`
}

func findRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, locationsFile)); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no %s above %s (use -root)", locationsFile, start)
		}
		dir = parent
	}
}

func readJSON(root, name string, v any) error {
	b, err := os.ReadFile(filepath.Join(root, name))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func loadGames(root string) (map[string]*Game, error) {
	var locs struct {
		Games map[string]struct {
			LocationPrefix string          `json:"locationPrefix"`
			Regions        json.RawMessage `json:"regions"`
		} `json:"games"`
	}
	if err := readJSON(root, locationsFile, &locs); err != nil {
		return nil, err
	}
	games := map[string]*Game{}
	for key, src := range locs.Games {
		g := &Game{Key: key, Prefix: src.LocationPrefix}
		var homeworlds []string
		types := map[string]string{}
		if key == "spyro1" {
			var s1 struct {
				BalloonistGates map[string]Gate `json:"balloonistGates"`
				Levels          []struct {
					Name string `json:"name"`
					Type string `json:"type"`
				} `json:"levels"`
			}
			if err := readJSON(root, fmt.Sprintf(dataFile, key), &s1); err != nil {
				return nil, err
			}
			for _, l := range s1.Levels {
				types[l.Name] = l.Type
				if l.Type == "HUB" {
					homeworlds = append(homeworlds, l.Name)
				}
			}
			g.rules = spyro1{gates: s1.BalloonistGates}
		} else {
			var d GateData
			if err := readJSON(root, fmt.Sprintf(dataFile, key), &d); err != nil {
				return nil, err
			}
			homeworlds = d.Homeworlds
			g.rules = newTabular(d)
		}
		if err := g.loadRegions(src.Regions, homeworlds); err != nil {
			return nil, fmt.Errorf("%s %s: %w", locationsFile, key, err)
		}
		for _, l := range g.Levels {
			l.Type = types[l.Name]
		}
		games[key] = g
	}
	return games, nil
}

// loadRegions keeps the regions in file order: Go maps drop it, and the order
// is what places each level under its homeworld.
func (g *Game) loadRegions(raw json.RawMessage, homeworlds []string) error {
	type region struct {
		GemTotal int     `json:"gemTotal"`
		Checks   []Check `json:"checks"`
		Events   []Check `json:"events"`
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	if _, err := dec.Token(); err != nil {
		return err
	}
	hubs := map[string]bool{}
	for _, h := range homeworlds {
		hubs[h] = true
	}
	var hub *Level
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		var r region
		if err := dec.Decode(&r); err != nil {
			return err
		}
		l := &Level{Name: tok.(string), Game: g, IsHub: hubs[tok.(string)], GemTotal: r.GemTotal}
		for _, e := range r.Events {
			e.Type = "event"
			r.Checks = append(r.Checks, e)
		}
		l.Checks = r.Checks
		if l.IsHub {
			l.Parent, hub = hub, l
		} else if hub == nil {
			return fmt.Errorf("level %q comes before any homeworld", l.Name)
		} else {
			l.Parent = hub
		}
		g.Levels = append(g.Levels, l)
	}
	if hub == nil {
		return fmt.Errorf("no homeworld found")
	}
	return nil
}
