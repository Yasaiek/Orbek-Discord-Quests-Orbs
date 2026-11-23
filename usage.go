package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// Usage is how much of the PC the program (with its windows) takes.
type Usage struct {
	CPU    float64 `json:"cpu"`    // % of the whole processor
	RAM    float64 `json:"ram"`    // MB
	RAMPct float64 `json:"ramPct"` // % of the PC's memory
	Procs  int     `json:"procs"`
	// the lowest and highest since the program started
	CPUMin float64 `json:"cpuMin"`
	CPUMax float64 `json:"cpuMax"`
	RAMMin float64 `json:"ramMin"`
	RAMMax float64 `json:"ramMax"`
	// the last few minutes, oldest first (one point every 2 s), for the little charts
	HistCPU []float64 `json:"histCpu"`
	HistRAM []float64 `json:"histRam"`
}

const histLen = 150 // 5 minutes

var (
	usageMu  sync.Mutex
	usageNow Usage
	// sums since usageAverage was last asked (for the anonymous statistics)
	avgCPU, avgRAM, avgPct float64
	avgN                   int
)

// usageAverage is the average use since the last call: CPU %, RAM MB, RAM % of the PC.
func usageAverage() (cpu, ram, pct float64) {
	usageMu.Lock()
	defer usageMu.Unlock()
	if avgN == 0 {
		u := usageNow
		return u.CPU, u.RAM, u.RAMPct
	}
	cpu, ram, pct = avgCPU/float64(avgN), avgRAM/float64(avgN), avgPct/float64(avgN)
	avgCPU, avgRAM, avgPct, avgN = 0, 0, 0, 0
	return
}

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
			if old := usageNow; old.RAM > 0 {
				u.CPUMin, u.CPUMax = math.Min(old.CPUMin, u.CPU), math.Max(old.CPUMax, u.CPU)
				u.RAMMin, u.RAMMax = math.Min(old.RAMMin, u.RAM), math.Max(old.RAMMax, u.RAM)
			} else {
				u.CPUMin, u.CPUMax, u.RAMMin, u.RAMMax = u.CPU, u.CPU, u.RAM, u.RAM
			}
			u.HistCPU = appendHist(usageNow.HistCPU, math.Round(u.CPU*100)/100)
			u.HistRAM = appendHist(usageNow.HistRAM, math.Round(u.RAM*10)/10)
			usageNow = u
			avgCPU, avgRAM, avgPct, avgN = avgCPU+u.CPU, avgRAM+u.RAM, avgPct+u.RAMPct, avgN+1
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

func appendHist(h []float64, v float64) []float64 {
	out := make([]float64, 0, histLen)
	if len(h) >= histLen {
		h = h[len(h)-histLen+1:]
	}
	return append(append(out, h...), v)
}
