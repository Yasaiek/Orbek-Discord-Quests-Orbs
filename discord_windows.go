//go:build windows

package main

import (
	"net"
	"strconv"
	"time"

	"github.com/Microsoft/go-winio"
)

// dialDiscord connects to the Discord client's local pipe number i.
func dialDiscord(i int) (net.Conn, error) {
	timeout := 2 * time.Second
	return winio.DialPipe(`\\.\pipe\discord-ipc-`+strconv.Itoa(i), &timeout)
}
