package main

import (
	"fmt"
	"path/filepath"
	"time"
)

// Rewards about to be lost: a quest is done but not claimed, and it ends within ExpiryHours. Orbek then
// shows a Windows notification (with its sound), plays a sound in the open panel and writes it in the
// journal – once per quest.

type expiryAlert struct {
	At   int64  `json:"at"`
	Text string `json:"text"`
}

func (e *Engine) expiryLoop() {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for {
		select {
		case <-e.quit:
			return
		case <-t.C:
		}
		e.checkExpiry()
	}
}

func (e *Engine) checkExpiry() {
	e.mu.Lock()
	if !e.set.ExpiryWarn || !e.connectedLocked() {
		e.mu.Unlock()
		return
	}
	if e.warned == nil {
		e.warned = map[string]int64{}
		readJSON(filepath.Join(e.dir, "warned.json"), &e.warned)
	}
	now := time.Now()
	within := time.Duration(max(1, e.set.ExpiryHours)) * time.Hour
	var notes []string
	for _, q := range e.last.Quests {
		if q.Completed == 0 || q.Claimed > 0 || q.Expires == 0 || e.warned[q.ID] > 0 {
			continue
		}
		left := time.UnixMilli(q.Expires).Sub(now)
		if left <= 0 || left > within {
			continue
		}
		e.warned[q.ID] = now.UnixMilli()
		h := int(left.Hours())
		when := fmt.Sprintf(tr("exp.hours"), h)
		if h < 1 {
			when = fmt.Sprintf(tr("exp.mins"), int(left.Minutes()))
		}
		notes = append(notes, fmt.Sprintf(tr("exp.one"), q.Name, when))
		e.log.Warn("L.expiring", q.Name, when)
	}
	// forget quests that are gone
	for id, at := range e.warned {
		if now.Sub(time.UnixMilli(at)) > 40*24*time.Hour {
			delete(e.warned, id)
		}
	}
	if len(notes) > 0 {
		writeJSON(filepath.Join(e.dir, "warned.json"), e.warned)
		text := notes[0]
		if len(notes) > 1 {
			text = fmt.Sprintf(tr("exp.many"), len(notes))
		}
		e.alert = expiryAlert{At: now.UnixMilli(), Text: text}
	}
	e.mu.Unlock()
	if len(notes) > 0 {
		notify(tr("exp.title"), e.alert.Text)
	}
}
