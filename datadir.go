package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// hasData: Orbek's files are already in this folder.
func hasData(dir string) bool {
	for _, n := range []string{"settings.json", "games.json"} {
		if _, err := os.Stat(filepath.Join(dir, n)); err == nil {
			return true
		}
	}
	return false
}

// legacyData brings the files over from where older versions kept them, next to the program file:
// "Orbek-data" (0.4.10–0.4.11) or "data" (up to 0.4.9; shared with Loorek when both sat in one folder).
func legacyData(dir string) {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	base := filepath.Dir(exe)
	if old := filepath.Join(base, "Orbek-data"); hasData(old) || dirExists(old) {
		if moveTree(old, dir) == nil {
			return
		}
	}
	moveOldData(filepath.Join(base, "data"), dir)
}

func dirExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

// moveOldData takes Orbek's files out of the old "data" folder. When Loorek lives there too, only what
// is Orbek's is taken and the rest is left alone.
func moveOldData(old, dir string) {
	if !dirExists(old) {
		return
	}
	shared := false
	for _, n := range []string{"auth.json", "login-profile", "hashes.json", "loorek-settings.json"} {
		if _, err := os.Stat(filepath.Join(old, n)); err == nil {
			shared = true
		}
	}
	if !shared {
		if moveTree(old, dir) == nil {
			return
		}
	}
	os.MkdirAll(dir, 0o700)
	if moveTree(filepath.Join(old, "tools"), filepath.Join(dir, "tools")) != nil {
		os.Rename(filepath.Join(old, "tools"), filepath.Join(dir, "tools"))
	}
	for _, n := range []string{"games.json", "warned.json"} {
		copyFile(filepath.Join(old, n), filepath.Join(dir, n))
	}
	if b, err := os.ReadFile(filepath.Join(old, "settings.json")); err == nil {
		var m map[string]any // only if the last one to write it was Orbek
		if json.Unmarshal(b, &m) == nil && (m["vencordDir"] != nil || m["refreshMin"] != nil) {
			os.WriteFile(filepath.Join(dir, "settings.json"), b, 0o600)
		}
	}
}
