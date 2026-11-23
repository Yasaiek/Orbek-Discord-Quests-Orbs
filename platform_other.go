//go:build !windows

package main

import (
	"os"
	"os/exec"
)

// On other systems (used for development and tests) there is no tray: just wait until quit.
func runTray(e *Engine, open func()) { <-e.quit }

func setAutostart(on bool) error { return nil }

func systemLocale() string {
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}

func hideWindow(cmd *exec.Cmd) {}
func tagWindows() {}
