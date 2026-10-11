// Author: Tim Harrison
// Created: 2026-09-26
// Updated: 2026-10-11

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/coder/websocket"
)

const apGame = "Spyro Reignited Trilogy"

type apClient struct {
	conn *websocket.Conn
	in   chan json.RawMessage // one server command per message
	errc chan error
}

type netItem struct {
	Item   int64 `json:"item"`
	Player int   `json:"player"`
}

type apPlayer struct {
	Slot  int    `json:"slot"`
	Alias string `json:"alias"`
}

// connect dials an Archipelago websocket server at host, defaulting the port to
// 38281 and trying wss:// before ws:// when no scheme is given.
func connect(ctx context.Context, host string) (*apClient, error) {
	if _, _, err := net.SplitHostPort(strings.TrimPrefix(strings.TrimPrefix(host, "wss://"), "ws://")); err != nil {
		host += ":38281"
	}
	urls := []string{host}
	if !strings.Contains(host, "://") {
		urls = []string{"wss://" + host, "ws://" + host}
	}
	var last error
	for _, u := range urls {
		dctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		conn, _, err := websocket.Dial(dctx, u, nil)
		cancel()
		if err == nil {
			conn.SetReadLimit(64 << 20)
			c := &apClient{conn: conn, in: make(chan json.RawMessage, 64), errc: make(chan error, 1)}
			go c.read(ctx)
			return c, nil
		}
		last = err
	}
	return nil, fmt.Errorf("can't reach %s: %w", host, last)
}

func (c *apClient) read(ctx context.Context) {
	for {
		_, b, err := c.conn.Read(ctx)
		if err != nil {
			c.errc <- err
			return
		}
		var cmds []json.RawMessage
		if err := json.Unmarshal(b, &cmds); err != nil {
			c.errc <- fmt.Errorf("bad server message: %w", err)
			return
		}
		for _, cmd := range cmds {
			select {
			case c.in <- cmd:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (c *apClient) send(ctx context.Context, cmds ...any) error {
	b, err := json.Marshal(cmds)
	if err != nil {
		return err
	}
	wctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return c.conn.Write(wctx, websocket.MessageText, b)
}

// cmdOf extracts the "cmd" field from a raw server command message,
// returning an empty string if it can't be decoded.
func cmdOf(raw json.RawMessage) string {
	var head struct {
		Cmd string `json:"cmd"`
	}
	_ = json.Unmarshal(raw, &head)
	return head.Cmd
}

// await skips server commands until one named cmd arrives.
func (c *apClient) await(ctx context.Context, cmd string) (json.RawMessage, error) {
	for {
		select {
		case raw := <-c.in:
			switch cmdOf(raw) {
			case cmd:
				return raw, nil
			case "ConnectionRefused":
				var p struct {
					Errors []string `json:"errors"`
				}
				_ = json.Unmarshal(raw, &p)
				return nil, fmt.Errorf("connection refused: %s", strings.Join(p.Errors, ", "))
			}
		case err := <-c.errc:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(30 * time.Second):
			return nil, fmt.Errorf("no %s from the server", cmd)
		}
	}
}

// itemNames fetches this game's item id -> name table.
func (c *apClient) itemNames(ctx context.Context) (map[int64]string, error) {
	if err := c.send(ctx, map[string]any{"cmd": "GetDataPackage", "games": []string{apGame}}); err != nil {
		return nil, err
	}
	raw, err := c.await(ctx, "DataPackage")
	if err != nil {
		return nil, err
	}
	var p struct {
		Data struct {
			Games map[string]struct {
				ItemNameToID map[string]int64 `json:"item_name_to_id"`
			} `json:"games"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	names := map[int64]string{}
	for name, id := range p.Data.Games[apGame].ItemNameToID {
		names[id] = name
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("the server's data package has no %s items", apGame)
	}
	return names, nil
}

type connected struct {
	Slot             int        `json:"slot"`
	Players          []apPlayer `json:"players"`
	CheckedLocations []int64    `json:"checked_locations"`
	MissingLocations []int64    `json:"missing_locations"`
	SlotData         SlotData   `json:"slot_data"`
}

func (c *apClient) connect(ctx context.Context, slot, password string) (connected, error) {
	var conn connected
	err := c.send(ctx, map[string]any{
		"cmd":            "Connect",
		"game":           apGame,
		"name":           slot,
		"password":       password,
		"uuid":           fmt.Sprintf("spyrosim-%x", time.Now().UnixNano()),
		"version":        map[string]any{"major": 0, "minor": 6, "build": 7, "class": "Version"},
		"items_handling": 7,
		"tags":           []string{},
		"slot_data":      true,
	})
	if err != nil {
		return conn, err
	}
	raw, err := c.await(ctx, "Connected")
	if err != nil {
		return conn, err
	}
	return conn, json.Unmarshal(raw, &conn)
}
