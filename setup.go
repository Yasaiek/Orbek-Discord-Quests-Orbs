package main

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// The one-click setup. Orbek's Discord side is a Vencord plugin, and a Vencord with your own plugins has to be
// built from its source. The setup does all of it for people who never touched that:
//   1. Node.js (a portable copy in Orbek's folder, nothing is installed system-wide)
//   2. the Vencord source (the user's own copy if there is one, otherwise downloaded)
//   3. the Orbek plugin put into it
//   4. the build (pnpm install + pnpm build)
//   5. Vencord put into the Discord app (Vencord's own installer, the same as "pnpm inject")
//   6. the plugin switched on in Vencord's settings, and Discord started again

type setupStep struct {
	Key   string `json:"key"`
	State string `json:"state"` // wait | run | ok | skip | fail
}

type setupState struct {
	mu      sync.Mutex
	Running bool        `json:"running"`
	Done    bool        `json:"done"`
	Err     string      `json:"err"`
	Steps   []setupStep `json:"steps"`
	Lines   []string    `json:"lines"`
	Full    bool        `json:"full"` // the whole setup, or only the plugin update
}

func (s *setupState) view() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	steps := append([]setupStep(nil), s.Steps...)
	lines := s.Lines
	if len(lines) > 40 {
		lines = lines[len(lines)-40:]
	}
	return map[string]any{"running": s.Running, "done": s.Done, "err": s.Err, "steps": steps, "lines": append([]string(nil), lines...), "full": s.Full}
}

func (s *setupState) line(format string, a ...any) {
	s.mu.Lock()
	s.Lines = append(s.Lines, strings.TrimSpace(fmt.Sprintf(format, a...)))
	if len(s.Lines) > 400 {
		s.Lines = s.Lines[len(s.Lines)-400:]
	}
	s.mu.Unlock()
}

func (s *setupState) mark(key, state string) {
	s.mu.Lock()
	for i := range s.Steps {
		if s.Steps[i].Key == key {
			s.Steps[i].State = state
		}
	}
	s.mu.Unlock()
}

// lineWriter turns a program's output into lines of the setup log.
type lineWriter struct {
	s   *setupState
	buf []byte
}

func (w *lineWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	for {
		i := strings.IndexAny(string(w.buf), "\r\n")
		if i < 0 {
			break
		}
		if l := strings.TrimSpace(stripANSI(string(w.buf[:i]))); l != "" {
			w.s.line("%s", l)
		}
		w.buf = w.buf[i+1:]
	}
	return len(p), nil
}

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b {
			for i++; i < len(s) && !((s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z')); i++ {
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func toolsDir(dir string) string { return filepath.Join(dir, "tools") }

// startSetup runs the setup in the background. full=false only refreshes the plugin (steps 3, 4, 6).
func (e *Engine) startSetup(full bool) bool {
	s := &e.setup
	s.mu.Lock()
	if s.Running {
		s.mu.Unlock()
		return false
	}
	keys := []string{"node", "vencord", "plugin", "build", "inject", "enable"}
	if !full {
		keys = []string{"node", "plugin", "build", "enable"}
	}
	s.Running, s.Done, s.Err, s.Lines, s.Full = true, false, "", nil, full
	s.Steps = nil
	for _, k := range keys {
		s.Steps = append(s.Steps, setupStep{Key: k, State: "wait"})
	}
	s.mu.Unlock()
	go func() {
		err := e.runSetup(full)
		s.mu.Lock()
		s.Running, s.Done = false, err == nil
		if err != nil {
			s.Err = err.Error()
			for i := range s.Steps {
				if s.Steps[i].State == "run" {
					s.Steps[i].State = "fail"
				}
			}
		}
		s.mu.Unlock()
		if err != nil {
			e.log.Warn("L.setupFail", err)
		} else {
			e.log.OK("L.setupDone")
		}
	}()
	return true
}

func (e *Engine) runSetup(full bool) error {
	s := &e.setup
	if runtime.GOOS != "windows" && os.Getenv("ORBEK_SETUP_TEST") == "" {
		return errors.New("the setup works on Windows only")
	}
	step := func(key string, fn func() error) error {
		s.mark(key, "run")
		if err := fn(); err != nil {
			s.line("✖ %v", err)
			return err
		}
		s.mark(key, "ok")
		return nil
	}
	var node string
	if err := step("node", func() (err error) { node, err = e.ensureNode(); return }); err != nil {
		return err
	}
	vdir := e.vencordDir()
	if full {
		if err := step("vencord", func() (err error) {
			if vdir != "" {
				s.line("Vencord: %s", vdir)
				return nil
			}
			vdir, err = e.fetchVencord()
			return
		}); err != nil {
			return err
		}
	} else if vdir == "" {
		return errors.New("no Vencord source folder found")
	}
	if err := step("plugin", func() error { return e.writePlugin(vdir) }); err != nil {
		return err
	}
	if err := step("build", func() error { return e.buildVencord(node, vdir) }); err != nil {
		return err
	}
	killDiscord()
	if full {
		if err := step("inject", func() error { return e.injectVencord(vdir) }); err != nil {
			return err
		}
	}
	if err := step("enable", enablePlugin); err != nil {
		return err
	}
	startDiscord()
	return nil
}

// ---- 1. Node.js ----

func (e *Engine) ensureNode() (string, error) {
	s := &e.setup
	nodeExe := "node"
	if runtime.GOOS == "windows" {
		nodeExe = "node.exe"
	}
	base := filepath.Join(toolsDir(e.dir), "node")
	if p := findIn(base, nodeExe); p != "" {
		s.line("Node.js: %s", p)
		return p, nil
	}
	if runtime.GOOS != "windows" { // tests
		if p, err := exec.LookPath("node"); err == nil {
			return p, nil
		}
	}
	s.line("Node.js: looking up the current LTS…")
	res, err := setupGet("https://nodejs.org/dist/index.json")
	if err != nil {
		return "", err
	}
	var list []struct {
		Version string   `json:"version"`
		Files   []string `json:"files"`
		LTS     any      `json:"lts"`
	}
	err = json.NewDecoder(io.LimitReader(res.Body, 8<<20)).Decode(&list)
	res.Body.Close()
	if err != nil {
		return "", err
	}
	ver := ""
	for _, v := range list { // newest first
		if lts, _ := v.LTS.(string); lts != "" && verNums(v.Version)[0] >= 22 && contains(v.Files, "win-x64-zip") {
			ver = v.Version
			break
		}
	}
	if ver == "" {
		return "", errors.New("no Node.js LTS found")
	}
	name := "node-" + ver + "-win-x64"
	zipPath := filepath.Join(toolsDir(e.dir), name+".zip")
	s.line("Node.js %s: downloading…", ver)
	if err := download("https://nodejs.org/dist/"+ver+"/"+name+".zip", zipPath); err != nil {
		return "", err
	}
	s.line("Node.js %s: unpacking…", ver)
	os.RemoveAll(base)
	if err := unzip(zipPath, base); err != nil {
		return "", err
	}
	os.Remove(zipPath)
	if p := findIn(base, nodeExe); p != "" {
		return p, nil
	}
	return "", errors.New("node.exe missing in the download")
}

// ---- 2. Vencord source ----

func (e *Engine) fetchVencord() (string, error) {
	s := &e.setup
	home, _ := os.UserHomeDir()
	dst := filepath.Join(home, "Documents", "Vencord")
	if _, err := os.Stat(dst); err == nil {
		dst = filepath.Join(home, "Documents", "Vencord-Orbek")
	}
	zipPath := filepath.Join(toolsDir(e.dir), "vencord.zip")
	s.line("Vencord: downloading the source…")
	if err := download("https://codeload.github.com/Vendicated/Vencord/zip/refs/heads/main", zipPath); err != nil {
		return "", err
	}
	tmp := dst + ".tmp"
	os.RemoveAll(tmp)
	if err := unzip(zipPath, tmp); err != nil {
		return "", err
	}
	os.Remove(zipPath)
	root := tmp // the archive holds one folder (Vencord-main)
	if ents, _ := os.ReadDir(tmp); len(ents) == 1 && ents[0].IsDir() {
		root = filepath.Join(tmp, ents[0].Name())
	}
	if err := os.Rename(root, dst); err != nil {
		return "", err
	}
	os.RemoveAll(tmp)
	e.mu.Lock()
	e.set.VencordDir = dst
	e.saveSettings()
	e.mu.Unlock()
	s.line("Vencord: %s", dst)
	return dst, nil
}

// ---- 3. plugin ----

func (e *Engine) writePlugin(vdir string) error {
	dst := filepath.Join(vdir, "src", "userplugins", pluginFolder)
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	for _, name := range []string{"index.tsx", "native.ts", "BUILD-I-RESTART.cmd"} {
		raw, err := pluginFS.ReadFile("plugin/" + name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dst, name), raw, 0o644); err != nil {
			return err
		}
	}
	e.setup.line("Plugin %s → %s", pluginVersion(), dst)
	return nil
}

// ---- 4. build ----

func (e *Engine) buildVencord(node, vdir string) error {
	s := &e.setup
	ver := "latest"
	var pkg struct {
		PM string `json:"packageManager"`
	}
	if readJSON(filepath.Join(vdir, "package.json"), &pkg) == nil && strings.HasPrefix(pkg.PM, "pnpm@") {
		ver = strings.SplitN(strings.TrimPrefix(pkg.PM, "pnpm@"), "+", 2)[0]
	}
	pnpmDir := filepath.Join(toolsDir(e.dir), "pnpm-"+ver)
	nodeDir := filepath.Dir(node)
	env := append(os.Environ(), "PATH="+nodeDir+string(os.PathListSeparator)+os.Getenv("PATH"), "CI=1", "npm_config_update_notifier=false")
	run := func(dir string, args ...string) error {
		cmd := exec.Command(node, args...)
		cmd.Dir, cmd.Env = dir, env
		w := &lineWriter{s: s}
		cmd.Stdout, cmd.Stderr = w, w
		hideWindow(cmd)
		return cmd.Run()
	}
	pnpm := pnpmEntry(pnpmDir)
	if pnpm == "" {
		npm := findIn(nodeDir, "npm-cli.js")
		if npm == "" { // a system Node keeps npm in lib/node_modules next to bin
			npm = findIn(filepath.Join(filepath.Dir(nodeDir), "lib"), "npm-cli.js")
		}
		if npm == "" {
			return errors.New("npm missing next to node")
		}
		s.line("pnpm %s: installing…", ver)
		os.RemoveAll(pnpmDir)
		os.MkdirAll(pnpmDir, 0o755)
		// No install scripts: newer npm blocks them anyway, and pnpm runs from its own Node entry
		// (bin/pnpm.mjs, which finds the native binary from the optional package).
		if err := run(pnpmDir, npm, "install", "--no-audit", "--no-fund", "--ignore-scripts", "--prefix", pnpmDir, "pnpm@"+ver); err != nil {
			return fmt.Errorf("pnpm: %w", err)
		}
		if pnpm = pnpmEntry(pnpmDir); pnpm == "" {
			return errors.New("pnpm: no entry point found after installing")
		}
	}
	s.line("pnpm install…")
	if err := run(vdir, pnpm, "install", "--frozen-lockfile"); err != nil {
		if err := run(vdir, pnpm, "install"); err != nil {
			return fmt.Errorf("pnpm install: %w", err)
		}
	}
	s.line("pnpm build…")
	if err := run(vdir, pnpm, "build"); err != nil {
		return fmt.Errorf("pnpm build: %w", err)
	}
	return nil
}

// pnpmEntry finds the file Node runs pnpm from: bin/pnpm.cjs up to pnpm 10, bin/pnpm.mjs from pnpm 11.
func pnpmEntry(dir string) string {
	base := filepath.Join(dir, "node_modules", "pnpm", "bin")
	for _, n := range []string{"pnpm.cjs", "pnpm.mjs", "pnpm.js"} {
		if _, err := os.Stat(filepath.Join(base, n)); err == nil {
			return filepath.Join(base, n)
		}
	}
	return ""
}

// ---- 5. into Discord ----

func (e *Engine) injectVencord(vdir string) error {
	s := &e.setup
	if runtime.GOOS != "windows" {
		return nil
	}
	cli := filepath.Join(vdir, "dist", "Installer", "VencordInstallerCli.exe")
	s.line("Vencord Installer: downloading…")
	if err := download("https://github.com/Vencord/Installer/releases/latest/download/VencordInstallerCli.exe", cli); err != nil {
		return err
	}
	try := func(args ...string) error {
		cmd := exec.Command(cli, args...)
		cmd.Dir = vdir
		cmd.Env = append(os.Environ(), "VENCORD_USER_DATA_DIR="+vdir, "VENCORD_DEV_INSTALL=1")
		w := &lineWriter{s: s}
		cmd.Stdout, cmd.Stderr = w, w
		hideWindow(cmd)
		return cmd.Run()
	}
	if err := try("--install", "--branch", "auto"); err != nil {
		s.line("retrying without --branch…")
		return try("--install")
	}
	return nil
}

// ---- 6. switch the plugin on ----

func vencordSettingsPath() string {
	base, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(base, "Vencord", "settings", "settings.json")
}

func enablePlugin() error {
	p := vencordSettingsPath()
	if p == "" {
		return errors.New("no settings folder")
	}
	m := map[string]any{}
	if raw, err := os.ReadFile(p); err == nil {
		if err := json.Unmarshal(raw, &m); err != nil {
			return fmt.Errorf("Vencord settings: %w", err)
		}
	}
	plugins, _ := m["plugins"].(map[string]any)
	if plugins == nil {
		plugins = map[string]any{}
	}
	for _, name := range []string{"Orbek", "VencordToolbox"} {
		pl, _ := plugins[name].(map[string]any)
		if pl == nil {
			pl = map[string]any{}
		}
		pl["enabled"] = true
		plugins[name] = pl
	}
	m["plugins"] = plugins
	raw, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return err
	}
	os.MkdirAll(filepath.Dir(p), 0o755)
	return os.WriteFile(p, raw, 0o644)
}

func killDiscord() {
	if runtime.GOOS != "windows" {
		return
	}
	for _, n := range []string{"Discord.exe", "DiscordPTB.exe", "DiscordCanary.exe"} {
		cmd := exec.Command("taskkill", "/F", "/IM", n)
		hideWindow(cmd)
		cmd.Run()
	}
	time.Sleep(2 * time.Second)
}

func startDiscord() {
	if runtime.GOOS != "windows" {
		return
	}
	upd := filepath.Join(os.Getenv("LOCALAPPDATA"), "Discord", "Update.exe")
	if _, err := os.Stat(upd); err == nil {
		exec.Command(upd, "--processStart", "Discord.exe").Start()
	}
}

// ---- helpers ----

func setupGet(u string) (*http.Response, error) {
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", appName+"/"+version)
	res, err := (&http.Client{Timeout: 10 * time.Minute}).Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		res.Body.Close()
		return nil, fmt.Errorf("%s: HTTP %d", u, res.StatusCode)
	}
	return res, nil
}

func download(u, dst string) error {
	res, err := setupGet(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	os.MkdirAll(filepath.Dir(dst), 0o755)
	f, err := os.Create(dst + ".part")
	if err != nil {
		return err
	}
	_, err = io.Copy(f, io.LimitReader(res.Body, 1<<30))
	f.Close()
	if err != nil {
		os.Remove(dst + ".part")
		return err
	}
	os.Remove(dst)
	return os.Rename(dst+".part", dst)
}

func unzip(src, dst string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	root := filepath.Clean(dst) + string(os.PathSeparator)
	for _, f := range r.File {
		p := filepath.Join(dst, filepath.FromSlash(f.Name))
		if !strings.HasPrefix(p, root) { // nothing outside the target folder
			continue
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(p, 0o755)
			continue
		}
		os.MkdirAll(filepath.Dir(p), 0o755)
		in, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode()|0o600)
		if err != nil {
			in.Close()
			return err
		}
		_, err = io.Copy(out, in)
		in.Close()
		out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// findIn looks for a file by name up to three folders deep.
func findIn(dir, name string) string {
	found := ""
	filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil || found != "" {
			return filepath.SkipDir
		}
		if rel, _ := filepath.Rel(dir, p); strings.Count(rel, string(os.PathSeparator)) > 4 {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.EqualFold(d.Name(), name) {
			found = p
		}
		return nil
	})
	return found
}

func contains(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}
