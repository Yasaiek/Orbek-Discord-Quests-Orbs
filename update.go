package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Updates come from the releases of the program's own GitHub repository: every release carries the program
// file and a small version.json with its version, size and SHA-256. A newer release is downloaded next to
// the running file, checked against that SHA-256, swapped in, and the program restarts itself in the tray.
const updateRepo = "Yasaiek/Orbek-Discord-Quests-Orbs"

// The updater takes its own copy of the program file from the release: GitHub counts every download of
// an asset, and updates should not inflate the download count of Orbek.exe shown on the project site.
const updateAsset = "Orbek-update.bin"

type updInfo struct {
	Version string `json:"version"`
	SHA256  string `json:"sha256"`
	Size    int64  `json:"size"`
}

type updState struct {
	State   string // "" | checking | current | available | downloading | failed
	Latest  string
	Checked time.Time
	Done    time.Time // when this copy started as a freshly installed update
	info    updInfo
	busy    bool
}

func updateBase() string {
	if runtime.GOOS != "windows" { // development and tests only
		if u := os.Getenv("ORBEK_UPDATE"); u != "" {
			return strings.TrimRight(u, "/")
		}
	}
	return "https://github.com/" + updateRepo
}

func trustedHost(h string) bool {
	h = strings.ToLower(h)
	return h == "github.com" || strings.HasSuffix(h, ".githubusercontent.com")
}

func verNums(v string) [3]int {
	var out [3]int
	for i, p := range strings.SplitN(strings.TrimPrefix(strings.TrimSpace(v), "v"), ".", 3) {
		n := 0
		for _, r := range p {
			if r < '0' || r > '9' {
				break
			}
			n = n*10 + int(r-'0')
		}
		out[i] = n
	}
	return out
}

func newer(a, b string) bool { // a > b
	x, y := verNums(a), verNums(b)
	for i := 0; i < 3; i++ {
		if x[i] != y[i] {
			return x[i] > y[i]
		}
	}
	return false
}

func updGet(u string, timeout time.Duration) (*http.Response, error) {
	hc := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", appName+"/"+version)
	res, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(u, "https://github.com/") && !trustedHost(res.Request.URL.Hostname()) {
		res.Body.Close()
		return nil, errors.New("unexpected download host " + res.Request.URL.Hostname())
	}
	return res, nil
}

func (e *Engine) updateLoop() {
	select {
	case <-e.quit:
		return
	case <-time.After(20 * time.Second):
	}
	for {
		e.updateCheck(false)
		select {
		case <-e.quit:
			return
		case <-time.After(6 * time.Hour):
		}
	}
}

// updateCheck looks for a newer release; with automatic updates on (or install=true) it installs it right away.
func (e *Engine) updateCheck(install bool) {
	e.mu.Lock()
	if e.upd.busy {
		e.mu.Unlock()
		return
	}
	e.upd.busy = true
	prev := e.upd.State
	e.upd.State = "checking"
	auto := e.set.AutoUpdate
	e.mu.Unlock()
	state, info := prev, updInfo{}
	res, err := updGet(updateBase()+"/releases/latest/download/version.json", 30*time.Second)
	switch {
	case err != nil:
		e.log.Debug("update check: %v", err)
		state = "failed"
	case res.StatusCode == 404: // nothing released yet
		res.Body.Close()
		state = "current"
	case res.StatusCode != 200:
		res.Body.Close()
		e.log.Debug("update check: HTTP %d", res.StatusCode)
		state = "failed"
	default:
		err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&info)
		res.Body.Close()
		if err != nil || info.Version == "" || len(info.SHA256) != 64 {
			e.log.Debug("update check: bad version.json (%v)", err)
			state = "failed"
		} else if newer(info.Version, version) {
			state = "available"
		} else {
			state = "current"
		}
	}
	e.mu.Lock()
	e.upd.State, e.upd.Checked = state, time.Now()
	if state == "available" {
		e.upd.Latest, e.upd.info = info.Version, info
	}
	e.upd.busy = false
	e.mu.Unlock()
	if state == "available" && (auto || install) {
		e.updateApply()
	}
}

func (e *Engine) updateApply() {
	e.mu.Lock()
	if e.upd.busy || e.upd.info.Version == "" {
		e.mu.Unlock()
		return
	}
	e.upd.busy = true
	e.upd.State = "downloading"
	info := e.upd.info
	e.mu.Unlock()
	e.log.Info("L.updFound", info.Version)
	err := e.updateInstall(info)
	e.mu.Lock()
	e.upd.busy = false
	if err != nil {
		e.upd.State = "failed"
	}
	e.mu.Unlock()
	if err != nil {
		e.log.Warn("L.updFail", err)
	}
}

func (e *Engine) updateInstall(info updInfo) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if p, err := filepath.EvalSymlinks(exe); err == nil {
		exe = p
	}
	res, err := updGet(updateBase()+"/releases/download/v"+info.Version+"/"+updateAsset, 15*time.Minute)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", res.StatusCode)
	}
	tmp, old := exe+".new", exe+".old"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(f, h), io.LimitReader(res.Body, 300<<20))
	f.Close()
	if err == nil && info.Size > 0 && n != info.Size {
		err = fmt.Errorf("size %d, expected %d", n, info.Size)
	}
	if err == nil && !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), info.SHA256) {
		err = errors.New("checksum mismatch")
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	// a running program file cannot be overwritten on Windows, but it can be renamed
	os.Remove(old)
	if err := os.Rename(exe, old); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, exe); err != nil {
		os.Rename(old, exe)
		os.Remove(tmp)
		return err
	}
	e.mu.Lock()
	gen := e.winGen
	e.mu.Unlock()
	// the new copy takes over this one's port and key, so that an open panel keeps working and reloads itself
	cmd := exec.Command(exe, "--tray", "--updated")
	cmd.Env = append(os.Environ(), "ORBEK_KEY="+e.apiKey, "ORBEK_GEN="+strconv.Itoa(gen), "ORBEK_PORT="+strconv.Itoa(e.port))
	if err := cmd.Start(); err != nil {
		os.Remove(exe)
		os.Rename(old, exe)
		return err
	}
	e.log.OK("L.updRestart", info.Version)
	go func() {
		time.Sleep(300 * time.Millisecond)
		if e.onQuit != nil {
			e.onQuit()
		}
	}()
	return nil
}

// cleanupUpdate removes what the previous version left behind after an update.
func cleanupUpdate() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	if p, err := filepath.EvalSymlinks(exe); err == nil {
		exe = p
	}
	os.Remove(exe + ".new")
	go func() {
		for i := 0; i < 10; i++ {
			if err := os.Remove(exe + ".old"); err == nil || os.IsNotExist(err) {
				return
			}
			time.Sleep(2 * time.Second)
		}
	}()
}
