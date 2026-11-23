package main

import (
	"fmt"
	"sync"
	"time"
)

// Usage is how much of the PC the program (with its windows) takes.
type Usage struct {
	CPU    float64 `json:"cpu"`    // % of the whole processor
	RAM    float64 `json:"ram"`    // MB
	RAMPct float64 `json:"ramPct"` // % of the PC's memory
	Procs  int     `json:"procs"`
}

var (
	usageMu  sync.Mutex
	usageNow Usage
)

func currentUsage() Usage {
	usageMu.Lock()
	defer usageMu.Unlock()
	return usageNow
}

// usageLoop measures every two seconds.
func usageLoop(quit <-chan struct{}) {
	s := &usageSampler{}
	s.sample()
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-quit:
			return
		case <-t.C:
			u := s.sample()
			usageMu.Lock()
			usageNow = u
			usageMu.Unlock()
		}
	}
}

// usageLine is the short text for the tray: "CPU 0.4% · RAM 61 MB (0.4%)".
func usageLine() string {
	u := currentUsage()
	return fmt.Sprintf("CPU %s%% · RAM %.0f MB (%s%%)", pct1(u.CPU), u.RAM, pct1(u.RAMPct))
}

func pct1(v float64) string {
	if v > 0 && v < 0.1 {
		return "<0.1"
	}
	return fmt.Sprintf("%.1f", v)
}
