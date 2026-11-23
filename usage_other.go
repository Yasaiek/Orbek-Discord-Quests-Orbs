//go:build !windows

package main

import "runtime"

func trackProcess(pid int) {}

type usageSampler struct{}

// Elsewhere (tests) just the Go program's own memory.
func (s *usageSampler) sample() Usage {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return Usage{RAM: float64(m.Sys) / (1 << 20), Procs: 1}
}
