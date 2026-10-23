// Author: Tim Harrison
// Created: 2026-09-26
// Updated: 2026-10-11

// spyrosim plays a Spyro Reignited Trilogy slot against an Archipelago
// server. It clears reachable levels, sends the checks checkd would send, and
// uses the items it receives.
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

// settle is how long the player waits for items before calling itself stuck,
// since its own checks send items back almost immediately.
const settle = 2 * time.Second

type config struct {
	server, slot, password, root string
	delay                        time.Duration
	exit                         bool
	dryRun                       bool
}

func main() {
	var cfg config
	flag.StringVar(&cfg.server, "server", "localhost:38281", "Archipelago server address")
	flag.StringVar(&cfg.slot, "slot", "", "slot name to play (required)")
	flag.StringVar(&cfg.password, "password", "", "room password")
	flag.StringVar(&cfg.root, "root", ".", "reignited-ap checkout, or any directory inside it")
	flag.DurationVar(&cfg.delay, "delay", time.Second, "time spent playing each level")
	flag.BoolVar(&cfg.exit, "exit", false, "exit when stuck instead of waiting: status 0 if the goal is done, 1 if not")
	flag.BoolVar(&cfg.dryRun, "dry-run", false, "never send LocationChecks — report reachability against a live slot without touching its real progress")
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

// run plays until ctx ends, or (with -exit) until the player is stuck, and
// reports whether the goal was reached.
func run(ctx context.Context, cfg config) (bool, error) {
	p, err := join(ctx, cfg)
	if err != nil {
		return false, err
	}
	defer func() { _ = p.client.conn.CloseNow() }()
	return p.play(ctx)
}

// player is one slot's run: the server link, the sim, and the state the event
// loop carries between messages.
type player struct {
	cfg       config
	client    *apClient
	sim       *Sim
	itemNames map[int64]string
	alias     map[int]string
	received  int
	goalSent  bool
	idleSince time.Time // When the player last ran out of reachable checks.
	stuck     bool
	wake      <-chan time.Time // Fires when the next level may be played.
}

// join connects to the server as the slot and builds the sim from its slot
// data and the locations the server already knows.
func join(ctx context.Context, cfg config) (_ *player, err error) {
	root, err := findRoot(cfg.root)
	if err != nil {
		return nil, err
	}
	games, err := loadGames(root)
	if err != nil {
		return nil, err
	}
	c, err := connect(ctx, cfg.server)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = c.conn.CloseNow()
		}
	}()
	if _, err := c.await(ctx, "RoomInfo"); err != nil {
		return nil, err
	}
	itemNames, err := c.itemNames(ctx)
	if err != nil {
		return nil, err
	}
	conn, err := c.connect(ctx, cfg.slot, cfg.password)
	if err != nil {
		return nil, err
	}
	sim, err := NewSim(games, conn.SlotData, append(conn.MissingLocations, conn.CheckedLocations...))
	if err != nil {
		return nil, err
	}
	sim.MarkChecked(conn.CheckedLocations)
	p := &player{
		cfg:       cfg,
		client:    c,
		sim:       sim,
		itemNames: itemNames,
		alias:     aliases(conn.Players),
		wake:      time.After(0),
	}
	log.Printf("connected as %s: %s, %d of %d locations checked", cfg.slot,
		strings.Join(conn.SlotData.IncludedGames, ", "), len(conn.CheckedLocations),
		len(conn.CheckedLocations)+len(conn.MissingLocations))
	return p, nil
}

// aliases maps every slot to its player name, with the server itself at 0.
func aliases(players []apPlayer) map[int]string {
	alias := map[int]string{0: "the server"}
	for _, p := range players {
		alias[p.Slot] = p.Alias
	}
	return alias
}

// play is the event loop: it applies server messages and plays the next
// reachable level whenever the play delay elapses.
func (p *player) play(ctx context.Context) (bool, error) {
	for {
		select {
		case <-ctx.Done():
			return p.goalSent, nil
		case err := <-p.client.errc:
			return p.goalSent, fmt.Errorf("lost the server: %w", err)
		case raw := <-p.client.in:
			if err := p.handle(raw); err != nil {
				return p.goalSent, err
			}
		case <-p.wake:
			if done, err := p.step(ctx); err != nil || done {
				return p.goalSent, err
			}
		}
	}
}

// handle applies one server message.
func (p *player) handle(raw json.RawMessage) error {
	switch cmdOf(raw) {
	case "ReceivedItems":
		return p.acceptAll(raw)
	case "RoomUpdate":
		p.markChecked(raw)
	}
	return nil
}

// acceptAll uses the items the server has newly sent, skipping any the run has
// already counted.
func (p *player) acceptAll(raw json.RawMessage) error {
	var m struct {
		Index int       `json:"index"`
		Items []netItem `json:"items"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}
	for i, it := range m.Items {
		if m.Index+i < p.received {
			continue
		}
		p.accept(it)
	}
	p.idleSince = time.Time{}
	if p.stuck {
		p.stuck = false
		p.wake = time.After(0)
	}
	return nil
}

// accept uses one received item and logs which player sent it.
func (p *player) accept(it netItem) {
	name, ok := p.itemNames[it.Item]
	if !ok {
		name = fmt.Sprintf("unknown item %d", it.Item)
	}
	p.sim.Receive(name)
	p.received++
	log.Printf("received %s from %s", name, p.alias[it.Player])
}

// markChecked records the locations the server reports as checked.
func (p *player) markChecked(raw json.RawMessage) {
	var m struct {
		CheckedLocations []int64 `json:"checked_locations"`
	}
	if err := json.Unmarshal(raw, &m); err == nil {
		p.sim.MarkChecked(m.CheckedLocations)
	}
}

// step plays the next reachable level, or waits out the settle delay when the
// player cannot get further. It reports whether the run should end.
func (p *player) step(ctx context.Context) (bool, error) {
	l := p.sim.Next()
	if l == nil {
		return p.wait()
	}
	p.idleSince = time.Time{}
	return false, p.playLevel(ctx, l)
}

// wait gives the server the settle delay to send items, then reports where the
// run is stuck. It reports whether -exit ends the run here.
func (p *player) wait() (bool, error) {
	if p.idleSince.IsZero() {
		p.idleSince = time.Now()
	}
	if time.Since(p.idleSince) < settle {
		p.wake = time.After(settle - time.Since(p.idleSince))
		return false, nil
	}
	p.stuck, p.wake = true, nil
	log.Printf("can't get further (%s); waiting for items:", p.sim.Progress())
	p.reportStuck()
	return p.cfg.exit, nil
}

// reportStuck logs the checks the player cannot reach and why.
func (p *player) reportStuck() {
	report := p.sim.Report()
	if len(report) == 0 {
		log.Printf("  nothing left: every location is checked")
	}
	for _, line := range report {
		log.Printf("  %s", line)
	}
}

// playLevel clears the level's reachable checks, sends them unless dry-run,
// and notes the goal. The next step waits the play delay.
func (p *player) playLevel(ctx context.Context, l *Level) error {
	got := p.sim.Play(l)
	verb := "played"
	if p.cfg.dryRun {
		p.sim.TakePending() // Dry-run never sends what it just marked played.
		verb = "reachable"
	} else if err := sendChecks(ctx, p.client, p.sim); err != nil {
		return err
	}
	log.Printf("%s %s%s: %s", verb, l.Game.Prefix, l.Name, describe(l, got))
	if err := p.noteGoal(ctx); err != nil {
		return err
	}
	p.wake = time.After(p.cfg.delay)
	return nil
}

// noteGoal sends the goal status once the slot's goal is done, or only logs it
// in dry-run.
func (p *player) noteGoal(ctx context.Context) error {
	if p.goalSent || !p.sim.GoalMet() {
		return nil
	}
	if p.cfg.dryRun {
		p.goalSent = true
		log.Printf("goal reachable (%s) — dry-run: not sending StatusUpdate", p.sim.Progress())
		return nil
	}
	if err := p.client.send(ctx, map[string]any{"cmd": "StatusUpdate", "status": 30}); err != nil {
		return err
	}
	p.goalSent = true
	log.Printf("goal complete (%s)", p.sim.Progress())
	return nil
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
