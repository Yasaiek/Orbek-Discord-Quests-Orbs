package main

import (
	"strconv"
	"strings"
	"time"
)

// Experimental claiming. The first try decides: when Discord refuses (a captcha, an error), claiming
// switches itself off and the panel says why.

// ClaimNote is how the last try went: idle | testing | ok | none (nothing to claim yet) | captcha | failed.
type ClaimNote struct {
	State  string `json:"state"`
	Reason string `json:"reason,omitempty"`
	At     int64  `json:"at"`
}

// claimResultLocked records a result from the plugin. Call with e.mu held.
func (e *Engine) claimResultLocked(state, reason string) {
	worked := e.set.ClaimNote.State == "ok"
	if worked && state == "failed" {
		return // claiming works; one quest Discord won't take (it is only in the journal) doesn't switch it off
	}
	e.set.ClaimNote = ClaimNote{State: state, Reason: reason, At: time.Now().UnixMilli()}
	if (state == "captcha" || state == "failed" || state == "learn" || state == "nofn") && e.set.AutoClaim {
		e.set.AutoClaim = false
		e.log.Warn("L.claimOff")
	}
	e.saveSettings()
}

// claimTest asks the plugin to try one claim now.
func (e *Engine) claimTest() {
	e.mu.Lock()
	at := time.Now().UnixMilli()
	switch {
	case !e.connectedLocked():
		e.set.ClaimNote = ClaimNote{State: "noplugin", At: at}
	case vnum(e.last.Plugin) < vnum("0.3.0"): // older plugins don't know the test
		e.set.ClaimNote = ClaimNote{State: "oldplugin", At: at}
	default:
		e.set.ClaimNote = ClaimNote{State: "testing", At: at}
	}
	e.saveSettings()
	testing := e.set.ClaimNote.State == "testing"
	e.mu.Unlock()
	if !testing {
		return
	}
	e.push(Cmd{Op: "claimTest"})
	time.AfterFunc(45*time.Second, func() { // no answer from the plugin
		e.mu.Lock()
		defer e.mu.Unlock()
		if e.set.ClaimNote.State == "testing" && e.set.ClaimNote.At == at {
			e.set.ClaimNote = ClaimNote{State: "noanswer", At: time.Now().UnixMilli()}
			e.saveSettings()
		}
	})
}

// vnum turns "0.3.10" into a comparable number.
func vnum(v string) int {
	n := 0
	for _, p := range strings.SplitN(v+".0.0", ".", 4)[:3] {
		x, _ := strconv.Atoi(p)
		n = n*1000 + x
	}
	return n
}
