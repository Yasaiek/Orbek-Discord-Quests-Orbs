package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Anonymous statistics for the website: how many people have the program running and how much of the PC it
// takes on average. Every 15 minutes: a random installation number (made here, tied to nothing), the
// version and the average CPU and RAM since the last time. Switched off in the settings, nothing is sent.

// statsURL is the statistics server (a Cloudflare Worker); "" = none yet.
const statsURL = ""

func (e *Engine) statsID() string {
	p := filepath.Join(e.dir, "stats-id")
	if b, err := os.ReadFile(p); err == nil && len(strings.TrimSpace(string(b))) == 32 {
		return strings.TrimSpace(string(b))
	}
	var r [16]byte
	rand.Read(r[:])
	id := hex.EncodeToString(r[:])
	os.WriteFile(p, []byte(id), 0o644)
	return id
}

func (e *Engine) statsLoop(app string) {
	if statsURL == "" {
		return
	}
	id := e.statsID()
	wait := 2 * time.Minute
	for {
		select {
		case <-e.quit:
			return
		case <-time.After(wait):
		}
		wait = 15 * time.Minute
		e.mu.Lock()
		on := e.set.Stats
		e.mu.Unlock()
		cpu, ram, ramPct := usageAverage()
		if !on {
			continue
		}
		raw, _ := json.Marshal(map[string]any{"app": app, "id": id, "v": version, "cpu": cpu, "ram": ram, "ramPct": ramPct})
		req, err := http.NewRequest("POST", strings.TrimRight(statsURL, "/")+"/ping", bytes.NewReader(raw))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		if res, err := (&http.Client{Timeout: 15 * time.Second}).Do(req); err == nil {
			res.Body.Close()
		}
	}
}
