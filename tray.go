package main

import (
	"fmt"
	"time"
)

const siteURL = "https://yasaiek.github.io/Orbek-Discord-Quests-Orbs/"

// trayActions are what the tray menu can do.
type trayActions struct {
	Open, Mini, StopAll, Update func()
}

type trayInfo struct {
	Line      string
	Connected bool
	Running   int
}

// trayState is the line under the program's name in the tray menu.
func (e *Engine) trayState() trayInfo {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.connectedLocked() {
		return trayInfo{Line: tr("tray.noDiscord")}
	}
	open := 0
	now := time.Now().UnixMilli()
	for i := range e.last.Quests {
		if e.last.Quests[i].open(now) {
			open++
		}
	}
	run := len(e.last.Runs)
	line := fmt.Sprintf(tr("tray.status"), open)
	if run > 0 {
		line += " · " + fmt.Sprintf(tr("tray.running"), run)
	}
	return trayInfo{Line: line, Connected: true, Running: run}
}
