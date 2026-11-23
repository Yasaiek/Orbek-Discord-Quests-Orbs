package main

import (
	"embed"
	"os"
	"path/filepath"
	"regexp"
)

// The Discord side of Orbek is a Vencord userplugin. Orbek carries its current files and can put them into
// the user's Vencord source (src/userplugins/orbQuests) and start the build, so the plugin updates with the app.

//go:embed plugin/index.tsx plugin/native.ts plugin/BUILD-I-RESTART.cmd
var pluginFS embed.FS

const pluginFolder = "orbQuests"

var rePluginVer = regexp.MustCompile(`PLUGIN_VERSION = "([0-9.]+)"`)

// pluginVersion: the version of the plugin files this program carries.
func pluginVersion() string {
	raw, _ := pluginFS.ReadFile("plugin/index.tsx")
	if m := rePluginVer.FindSubmatch(raw); m != nil {
		return string(m[1])
	}
	return ""
}

// vencordDir finds the Vencord source folder: the one in the settings, or the usual places.
func (e *Engine) vencordDir() string {
	e.mu.Lock()
	pref := e.set.VencordDir
	e.mu.Unlock()
	home, _ := os.UserHomeDir()
	for _, d := range []string{pref, filepath.Join(home, "Documents", "Vencord"), filepath.Join(home, "Vencord"),
		filepath.Join(home, "Desktop", "Vencord"), filepath.Join(home, "Downloads", "Vencord")} {
		if d == "" {
			continue
		}
		if st, err := os.Stat(filepath.Join(d, "src", "userplugins")); err == nil && st.IsDir() {
			return d
		}
		if st, err := os.Stat(filepath.Join(d, "src", "plugins")); err == nil && st.IsDir() {
			return d // no userplugins folder yet: it is made below
		}
	}
	return ""
}
