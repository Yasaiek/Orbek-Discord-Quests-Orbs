//go:build !windows

package main

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// dialDiscord connects to the Discord client's local socket number i.
func dialDiscord(i int) (net.Conn, error) {
	dir := os.Getenv("DROPOMAT_DISCORD") // tests point this at a stand-in
	for _, k := range []string{"XDG_RUNTIME_DIR", "TMPDIR"} {
		if dir == "" {
			dir = os.Getenv(k)
		}
	}
	if dir == "" {
		dir = "/tmp"
	}
	return net.DialTimeout("unix", filepath.Join(dir, "discord-ipc-"+strconv.Itoa(i)), 2*time.Second)
}
