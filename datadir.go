package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

// Up to 0.4.9 Orbek kept its files in "data" next to the executable. Loorek does the same, so with both
// programs in one folder they shared settings.json and overwrote each other's settings. Orbek now has a
// folder of its own; moveOldData moves what is Orbek's out of the old one, once.
func moveOldData(old, dir string) {
	if _, err := os.Stat(dir); err == nil {
		return // already moved
	}
	if _, err := os.Stat(old); err != nil {
		return // first start
	}
	shared := false // Loorek lives there too: take only Orbek's files, leave the rest
	for _, n := range []string{"auth.json", "login-profile", "hashes.json", "loorek-settings.json"} {
		if _, err := os.Stat(filepath.Join(old, n)); err == nil {
			shared = true
		}
	}
	if !shared {
		if os.Rename(old, dir) == nil {
			return
		}
	}
	os.MkdirAll(dir, 0o700)
	os.Rename(filepath.Join(old, "tools"), filepath.Join(dir, "tools"))
	for _, n := range []string{"games.json", "warned.json", "stats-id"} {
		copyFile(filepath.Join(old, n), filepath.Join(dir, n))
	}
	if b, err := os.ReadFile(filepath.Join(old, "settings.json")); err == nil {
		var m map[string]any // only if the last one to write it was Orbek
		if json.Unmarshal(b, &m) == nil && (m["vencordDir"] != nil || m["refreshMin"] != nil) {
			os.WriteFile(filepath.Join(dir, "settings.json"), b, 0o600)
		}
	}
}

func copyFile(src, dst string) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	io.Copy(out, in)
	out.Close()
}
