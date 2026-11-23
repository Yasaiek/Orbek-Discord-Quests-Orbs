package main

import "time"

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
	e.set.ClaimNote = ClaimNote{State: state, Reason: reason, At: time.Now().UnixMilli()}
	if (state == "captcha" || state == "failed") && e.set.AutoClaim {
		e.set.AutoClaim = false
		e.log.Warn("L.claimOff")
	}
	e.saveSettings()
}

// claimTest asks the plugin to try one claim now.
func (e *Engine) claimTest() {
	e.mu.Lock()
	e.set.ClaimNote = ClaimNote{State: "testing", At: time.Now().UnixMilli()}
	e.saveSettings()
	e.mu.Unlock()
	e.push(Cmd{Op: "claimTest"})
}
