package main

import (
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// What the automat may do by itself: kinds of reward, kinds of task and which games.

var allKinds = []string{"orbs", "cosmetic", "ingame", "code", "other"}
var allTasks = []string{"video", "play", "activity", "stream"}

func rewardKind(t int) string {
	switch t {
	case 4:
		return "orbs"
	case 3:
		return "cosmetic"
	case 2:
		return "ingame"
	case 1:
		return "code"
	}
	return "other"
}

func taskKind(task string) string {
	switch {
	case strings.HasPrefix(task, "WATCH_VIDEO"):
		return "video"
	case task == "STREAM_ON_DESKTOP":
		return "stream"
	case task == "PLAY_ACTIVITY" || task == "ACHIEVEMENT_IN_ACTIVITY":
		return "activity"
	}
	return "play"
}

func has(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// migrateFilters fills the new settings from the old "autoOnly" choice.
func (s *Settings) migrateFilters() {
	if s.AutoKinds == nil {
		switch s.AutoOnly {
		case "orbs":
			s.AutoKinds = []string{"orbs"}
		case "cosmetic":
			s.AutoKinds = []string{"cosmetic"}
		default:
			s.AutoKinds = append([]string{}, allKinds...)
		}
	}
	if s.AutoTasks == nil {
		s.AutoTasks = []string{"video", "play", "activity"} // a stream needs a voice channel with someone
	}
	if s.AutoGames == "" {
		s.AutoGames = "all"
	}
	if s.GameList == nil {
		s.GameList = []string{}
	}
}

// autoWants: may the automat take this quest by itself?
func (s *Settings) autoWants(q *Quest) bool {
	if !has(s.AutoTasks, taskKind(q.Task)) {
		return false
	}
	kind := false
	for _, r := range q.Rewards {
		if has(s.AutoKinds, rewardKind(r.Type)) {
			kind = true
		}
	}
	if !kind {
		return false
	}
	switch s.AutoGames {
	case "only":
		return has(s.GameList, q.Game)
	case "except":
		return !has(s.GameList, q.Game)
	}
	return true
}

// cleanList keeps only known values, in their usual order.
func cleanList(v any, known []string) []string {
	out := []string{}
	if arr, ok := v.([]any); ok {
		for _, k := range known {
			for _, x := range arr {
				if str(x) == k {
					out = append(out, k)
					break
				}
			}
		}
	}
	return out
}

// ---- games seen so far (to choose from in the settings) ----

type seenGame struct {
	Name string `json:"name"`
	Last int64  `json:"last"` // when a quest of it was last seen
	Icon string `json:"icon,omitempty"`
}

func (e *Engine) loadGames() {
	var list []seenGame
	readJSON(filepath.Join(e.dir, "games.json"), &list)
	e.games, e.gameIcons = map[string]int64{}, map[string]string{}
	for _, g := range list {
		e.games[g.Name] = g.Last
		if g.Icon != "" {
			e.gameIcons[g.Name] = g.Icon
		}
	}
}

// noteGamesLocked remembers the games of the quests in the list (saved at most every minute).
func (e *Engine) noteGamesLocked(qs []Quest) {
	if e.games == nil {
		e.loadGames()
	}
	now := time.Now().UnixMilli()
	changed := false
	for i := range qs {
		g := strings.TrimSpace(qs[i].Game)
		if g == "" {
			continue
		}
		if _, ok := e.games[g]; !ok {
			changed = true
		}
		e.games[g] = now
		if ic := qs[i].Icon; ic != "" && e.gameIcons[g] != ic {
			e.gameIcons[g] = ic
			changed = true
		}
	}
	if changed || time.Since(e.gamesSaved) > time.Hour {
		e.gamesSaved = time.Now()
		writeJSON(filepath.Join(e.dir, "games.json"), e.gameListLocked())
	}
}

func (e *Engine) gameListLocked() []seenGame {
	out := make([]seenGame, 0, len(e.games))
	for n, t := range e.games {
		out = append(out, seenGame{n, t, e.gameIcons[n]})
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out
}
