//go:build windows

package main

import (
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// How much of the PC Orbek takes: this program plus its windows (the browser processes it started and
// everything they started). CPU is a share of the whole processor (all cores), RAM the working set.

var (
	launchedMu sync.Mutex
	launched   = map[uint32]bool{}

	procMemInfo   = syscall.NewLazyDLL("kernel32.dll").NewProc("K32GetProcessMemoryInfo")
	procMemStatus = syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx")
)

// trackProcess remembers a process Orbek started (a window), so its use is counted.
func trackProcess(pid int) {
	launchedMu.Lock()
	launched[uint32(pid)] = true
	launchedMu.Unlock()
}

type procMemCounters struct {
	cb                         uint32
	pageFaultCount             uint32
	peakWorkingSetSize         uintptr
	workingSetSize             uintptr
	quotaPeakPagedPoolUsage    uintptr
	quotaPagedPoolUsage        uintptr
	quotaPeakNonPagedPoolUsage uintptr
	quotaNonPagedPoolUsage     uintptr
	pagefileUsage              uintptr
	peakPagefileUsage          uintptr
}

type memStatusEx struct {
	length, memoryLoad                 uint32
	totalPhys, availPhys               uint64
	totalPageFile, availPageFile       uint64
	totalVirtual, availVirtual, extVir uint64
}

type usageSampler struct {
	last  map[uint32]int64 // CPU time (100 ns) per process at the last sample
	lastT time.Time
}

// ourProcesses: this program, the windows it started and all their children.
func ourProcesses() []uint32 {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return []uint32{uint32(windows.GetCurrentProcessId())}
	}
	defer windows.CloseHandle(snap)
	parent := map[uint32]uint32{}
	alive := map[uint32]bool{}
	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	for err = windows.Process32First(snap, &pe); err == nil; err = windows.Process32Next(snap, &pe) {
		parent[pe.ProcessID] = pe.ParentProcessID
		alive[pe.ProcessID] = true
	}
	roots := map[uint32]bool{uint32(windows.GetCurrentProcessId()): true}
	launchedMu.Lock()
	for pid := range launched {
		if alive[pid] {
			roots[pid] = true
		} else {
			delete(launched, pid)
		}
	}
	launchedMu.Unlock()
	var out []uint32
	for pid := range alive {
		for p, n := pid, 0; n < 12; p, n = parent[p], n+1 {
			if roots[p] {
				out = append(out, pid)
				break
			}
			if parent[p] == p || parent[p] == 0 {
				break
			}
		}
	}
	return out
}

func (s *usageSampler) sample() Usage {
	now := time.Now()
	cur := map[uint32]int64{}
	var cpuDelta int64
	var ram uint64
	for _, pid := range ourProcesses() {
		h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
		if err != nil {
			continue
		}
		var c, x, k, u windows.Filetime
		if windows.GetProcessTimes(h, &c, &x, &k, &u) == nil {
			t := int64(k.HighDateTime)<<32 | int64(k.LowDateTime) + (int64(u.HighDateTime)<<32 | int64(u.LowDateTime))
			cur[pid] = t
			if old, ok := s.last[pid]; ok && t >= old {
				cpuDelta += t - old
			}
		}
		var m procMemCounters
		m.cb = uint32(unsafe.Sizeof(m))
		if r, _, _ := procMemInfo.Call(uintptr(h), uintptr(unsafe.Pointer(&m)), uintptr(m.cb)); r != 0 {
			ram += uint64(m.workingSetSize)
		}
		windows.CloseHandle(h)
	}
	u := Usage{RAM: float64(ram) / (1 << 20), Procs: len(cur)}
	if !s.lastT.IsZero() {
		wall := now.Sub(s.lastT).Nanoseconds() / 100 // in 100 ns, like the CPU times
		if wall > 0 {
			u.CPU = float64(cpuDelta) / float64(wall*int64(runtime.NumCPU())) * 100
		}
	}
	var ms memStatusEx
	ms.length = uint32(unsafe.Sizeof(ms))
	if r, _, _ := procMemStatus.Call(uintptr(unsafe.Pointer(&ms))); r != 0 && ms.totalPhys > 0 {
		u.RAMPct = float64(ram) / float64(ms.totalPhys) * 100
	}
	s.last, s.lastT = cur, now
	return u
}
