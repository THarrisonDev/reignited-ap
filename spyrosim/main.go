// spyrosim plays a Spyro Reignited Trilogy slot against an Archipelago server:
// it clears whatever levels the simulated player can reach, sends the checks
// checkd would send, uses the items it receives, and waits when it can't get
// further.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

// settle is how long the player waits for items before calling itself stuck:
// items found by its own checks come back a moment after they are sent.
const settle = 2 * time.Second

type config struct {
	server, slot, password, root string
	delay                        time.Duration
	exit                         bool
}

func main() {
	var cfg config
	flag.StringVar(&cfg.server, "server", "localhost:38281", "Archipelago server address")
	flag.StringVar(&cfg.slot, "slot", "", "slot name to play (required)")
	flag.StringVar(&cfg.password, "password", "", "room password")
	flag.StringVar(&cfg.root, "root", ".", "reignited-ap checkout, or any directory inside it")
	flag.DurationVar(&cfg.delay, "delay", time.Second, "time spent playing each level")
	flag.BoolVar(&cfg.exit, "exit", false, "exit when stuck instead of waiting: status 0 if the goal is done, 1 if not")
	flag.Parse()
	if cfg.slot == "" {
		flag.Usage()
		os.Exit(2)
	}
	log.SetFlags(log.Ltime)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	done, err := run(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	if !done {
		os.Exit(1)
	}
}

// run plays until ctx ends, or (with -exit) until stuck; it reports whether
// the goal was reached.
func run(ctx context.Context, cfg config) (bool, error) {
	root, err := findRoot(cfg.root)
	if err != nil {
		return false, err
	}
	games, err := loadGames(root)
	if err != nil {
		return false, err
	}
	c, err := connect(ctx, cfg.server)
	if err != nil {
		return false, err
	}
	defer c.conn.CloseNow()
	if _, err := c.await(ctx, "RoomInfo"); err != nil {
		return false, err
	}
	itemNames, err := c.itemNames(ctx)
	if err != nil {
		return false, err
	}
	conn, err := c.connect(ctx, cfg.slot, cfg.password)
	if err != nil {
		return false, err
	}
	sim, err := NewSim(games, conn.SlotData, append(conn.MissingLocations, conn.CheckedLocations...))
	if err != nil {
		return false, err
	}
	sim.MarkChecked(conn.CheckedLocations)
	alias := map[int]string{0: "the server"}
	for _, p := range conn.Players {
		alias[p.Slot] = p.Alias
	}
	log.Printf("connected as %s: %s, %d of %d locations checked", cfg.slot,
		strings.Join(conn.SlotData.IncludedGames, ", "), len(conn.CheckedLocations),
		len(conn.CheckedLocations)+len(conn.MissingLocations))

	received := 0
	goalSent := false
	idleSince := time.Time{}
	stuck := false
	wake := time.After(0)
	for {
		select {
		case <-ctx.Done():
			return goalSent, nil
		case err := <-c.errc:
			return goalSent, fmt.Errorf("lost the server: %w", err)

		case raw := <-c.in:
			switch cmdOf(raw) {
			case "ReceivedItems":
				var p struct {
					Index int       `json:"index"`
					Items []netItem `json:"items"`
				}
				if err := json.Unmarshal(raw, &p); err != nil {
					return goalSent, err
				}
				for i, it := range p.Items {
					if p.Index+i < received {
						continue
					}
					name, ok := itemNames[it.Item]
					if !ok {
						name = fmt.Sprintf("unknown item %d", it.Item)
					}
					sim.Receive(name)
					received++
					log.Printf("received %s from %s", name, alias[it.Player])
				}
				idleSince = time.Time{}
				if stuck {
					stuck = false
					wake = time.After(0)
				}
			case "RoomUpdate":
				var p struct {
					CheckedLocations []int64 `json:"checked_locations"`
				}
				if err := json.Unmarshal(raw, &p); err == nil {
					sim.MarkChecked(p.CheckedLocations)
				}
			}

		case <-wake:
			l := sim.Next()
			if l == nil {
				if idleSince.IsZero() {
					idleSince = time.Now()
				}
				if time.Since(idleSince) < settle {
					wake = time.After(settle - time.Since(idleSince))
					continue
				}
				stuck, wake = true, nil
				log.Printf("can't get further (%s); waiting for items:", sim.Progress())
				report := sim.Report()
				if len(report) == 0 {
					log.Printf("  nothing left: every location is checked")
				}
				for _, line := range report {
					log.Printf("  %s", line)
				}
				if cfg.exit {
					return goalSent, nil
				}
				continue
			}
			idleSince = time.Time{}
			got := sim.Play(l)
			if err := sendChecks(ctx, c, sim); err != nil {
				return goalSent, err
			}
			log.Printf("played %s%s: %s", l.Game.Prefix, l.Name, describe(l, got))
			if !goalSent && sim.GoalMet() {
				if err := c.send(ctx, map[string]any{"cmd": "StatusUpdate", "status": 30}); err != nil {
					return goalSent, err
				}
				goalSent = true
				log.Printf("goal complete (%s)", sim.Progress())
			}
			wake = time.After(cfg.delay)
		}
	}
}

func sendChecks(ctx context.Context, c *apClient, sim *Sim) error {
	ids := sim.TakePending()
	if len(ids) == 0 {
		return nil
	}
	return c.send(ctx, map[string]any{"cmd": "LocationChecks", "locations": ids})
}

func describe(l *Level, got []string) string {
	if len(got) == 0 {
		return "nothing to send"
	}
	short := make([]string, len(got))
	for i, name := range got {
		short[i] = strings.TrimPrefix(strings.TrimPrefix(name, l.Game.Prefix), l.Name+": ")
	}
	return strings.Join(short, ", ")
}
