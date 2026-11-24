package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

//go:embed ui
var uiFS embed.FS

type server struct {
	e    *Engine
	key  string
	port int
}

func localHost(r *http.Request) bool {
	host := r.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return host == "127.0.0.1" || host == "localhost"
}

func (s *server) guard(w http.ResponseWriter, r *http.Request, header string) bool {
	if !localHost(r) || r.Header.Get(header) != s.key {
		http.Error(w, "forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func (e *Engine) view() map[string]any {
	vencord := e.vencordDir()
	e.mu.Lock()
	defer e.mu.Unlock()
	s := e.last
	quests := make([]map[string]any, 0, len(s.Quests))
	for i := range s.Quests {
		q := &s.Quests[i]
		quests = append(quests, map[string]any{
			"id": q.ID, "name": q.Name, "game": q.Game, "publisher": q.Publisher, "task": q.Task,
			"target": q.Target, "progress": q.Progress, "rewards": q.Rewards, "hero": q.Hero, "icon": q.Icon,
			"startsAt": q.Starts, "expiresAt": q.Expires, "enrolledAt": q.Enrolled, "completedAt": q.Completed, "claimedAt": q.Claimed,
			"auto": q.Auto, "orbs": q.orbs(), "cosmetic": q.cosmetic(),
			"taskTitle": q.TaskTitle, "taskDesc": q.TaskDesc, "actName": q.ActName, "actUrl": q.ActURL,
		})
	}
	queue := s.Queue
	if queue == nil {
		queue = []string{}
	}
	runs := s.Runs
	if runs == nil {
		runs = []RunState{}
	}
	return map[string]any{
		"app": appName, "closing": e.closing, "dataDir": e.dir, "dataDefault": dataDirIn(defaultRoot()), "dataCustom": readLocation().Root != "", "version": version, "lang": curLang.Load(), "winGen": e.winGen,
		"connected": e.connectedLocked(), "lastAt": ms(e.lastAt), "user": s.User, "loaded": s.Loaded, "err": s.Err,
		"orbs": s.Orbs, "runs": runs, "queue": queue, "desktop": s.Desktop, "plugin": s.Plugin, "pluginWait": e.plugWaitLocked(), "pluginStuck": e.plugStuckLocked(),
		"pluginLatest": pluginVersion(), "vencord": vencord, "usage": currentUsage(), "games": e.gameListLocked(), "rpc": map[string]any{"state": e.rpcState, "ready": discordAppID != ""}, "statsOn": statsURL != "", "alert": e.alert, "orbTile": e.set.OrbTile, "orbTiles": e.set.OrbTiles, "orbGlyph": e.set.OrbGlyph,
		"setup":    e.setup.view(),
		"update":   map[string]any{"state": e.upd.State, "latest": e.upd.Latest, "checked": ms(e.upd.Checked), "doneAt": ms(e.upd.Done), "from": e.upd.From, "got": updGot.Load(), "size": updSize.Load()},
		"quests":   quests,
		"settings": e.set,
		"log":      e.log.tail(),
	}
}

func ms(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixMilli()
}

type opReq struct {
	Op    string         `json:"op"`
	ID    string         `json:"id"`
	IDs   []string       `json:"ids"`
	Patch map[string]any `json:"patch"`
}

func (s *server) handleOp(r opReq) map[string]any {
	e := s.e
	switch r.Op {
	case "refresh", "stopAll":
		e.push(Cmd{Op: r.Op})
	case "claimAll": // every reward waiting to be claimed, soonest ending first: Discord shows them one by one
		e.mu.Lock()
		now := time.Now().UnixMilli()
		var w []Quest
		for _, q := range e.last.Quests {
			if q.Completed > 0 && q.Claimed == 0 && (q.Expires == 0 || q.Expires > now) {
				w = append(w, q)
			}
		}
		e.mu.Unlock()
		sort.Slice(w, func(i, j int) bool { return w[i].Expires < w[j].Expires })
		ids := make([]string, len(w))
		for i := range w {
			ids[i] = w[i].ID
		}
		e.push(Cmd{Op: "claim", IDs: ids})
	case "enroll", "stop", "claim", "show":
		if r.ID == "" {
			return map[string]any{"ok": false}
		}
		e.push(Cmd{Op: r.Op, ID: r.ID})
	case "run":
		e.mu.Lock()
		allowed := e.set.AutoMode
		for _, id := range r.IDs {
			delete(e.fails, id)
		}
		e.mu.Unlock()
		if !allowed || len(r.IDs) == 0 {
			return map[string]any{"ok": false}
		}
		e.push(Cmd{Op: "run", IDs: r.IDs})
	case "settings":
		e.mu.Lock()
		langChanged, autostart := false, (*bool)(nil)
		for k, v := range r.Patch {
			switch k {
			case "lang":
				e.set.Lang, langChanged = str(v), true
			case "autoMode":
				e.set.AutoMode = boolean(v)
			case "autoAll":
				e.set.AutoAll = boolean(v)
			case "autoKinds":
				e.set.AutoKinds = cleanList(v, allKinds)
			case "autoTasks":
				e.set.AutoTasks = cleanList(v, allTasks)
			case "autoGames":
				if x := str(v); x == "all" || x == "only" || x == "except" {
					e.set.AutoGames = x
				}
			case "gameList":
				list := []string{}
				if arr, ok := v.([]any); ok {
					for _, x := range arr {
						if s := strings.TrimSpace(str(x)); s != "" && !has(list, s) {
							list = append(list, s)
						}
					}
				}
				e.set.GameList = list
			case "openOnStart":
				e.set.OpenOnStart = boolean(v)
			case "autostart":
				b := boolean(v)
				e.set.Autostart, autostart = b, &b
			case "parallel":
				if n := int(num(v)); n >= 0 && n <= 50 {
					e.set.Parallel = n
				}
			case "autoUpdate":
				e.set.AutoUpdate = boolean(v)
			case "clickSelf":
				e.set.ClickSelf = boolean(v)
			case "expiryWarn":
				e.set.ExpiryWarn = boolean(v)
			case "expiryHours":
				if n := int(num(v)); n >= 1 && n <= 168 {
					e.set.ExpiryHours = n
				}
			case "stats":
				e.set.Stats = boolean(v)
			case "orbAnim":
				e.set.OrbAnim = boolean(v)
			case "rpcOn":
				e.set.RpcOn = boolean(v)
			case "rpcQuest":
				e.set.RpcQuest = boolean(v)
			case "rpcOrbs":
				e.set.RpcOrbs = boolean(v)
			case "rpcToGet":
				e.set.RpcToGet = boolean(v)
			case "rpcBtn":
				e.set.RpcBtn = boolean(v)
			case "refreshMin":
				if n := int(num(v)); n >= 1 && n <= 43200 {
					e.set.RefreshMin = n
				}
			}
		}
		off := !e.set.AutoMode
		e.saveSettings()
		e.mu.Unlock()

		if off {
			e.push(Cmd{Op: "stopAll"})
		}
		if langChanged {
			e.applyLang()
		}
		if autostart != nil {
			if err := setAutostart(*autostart); err != nil {
				e.log.Warn("L.autostartFail", err)
			}
		}
	case "clearLog":
		e.log.clear()
	case "pluginInstall":
		e.mu.Lock()
		current := e.connectedLocked() && e.last.Plugin != "" && !newer(pluginVersion(), e.last.Plugin)
		e.mu.Unlock()
		if current { // Discord already runs the newest plugin
			return map[string]any{"ok": false, "err": "current"}
		}
		return map[string]any{"ok": e.startSetup(false)}
	case "setup":
		return map[string]any{"ok": e.startSetup(true)}
	case "updCheck":
		go e.updateCheck(false)
	case "updInstall":
		go e.updateCheck(true)
	case "mini":
		go s.showMini()
	case "quit":
		go e.quitApp()
	case "dataOpen":
		openFolder(e.dir)
	case "dataPick": // the system folder picker; the program's own folder is made inside the one picked
		p := pickFolder(tr("set.dataDir"), filepath.Dir(e.dir))
		if p == "" {
			return map[string]any{"ok": false}
		}
		return map[string]any{"ok": true, "path": p, "target": dataDirIn(p)}
	case "dataMove":
		if os.Getenv("ORBEK_DATA") != "" {
			return map[string]any{"ok": false, "err": "bad"}
		}
		dir, err := chooseRoot(r.ID, e.dir)
		if err != nil {
			return map[string]any{"ok": false, "err": err.Error()}
		}
		e.log.Info("L.dataMoving", dir)
		go func() {
			time.Sleep(400 * time.Millisecond)
			if err := e.restartSelf("--moved"); err != nil {
				saveLocation(readLocation()) // nothing moved: the next start does it
				e.log.Err("L.dataMoveFail", err)
			}
		}()
		return map[string]any{"ok": true, "target": dir}
	default:
		return map[string]any{"ok": false}
	}
	return map[string]any{"ok": true}
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		if !s.guard(w, r, "X-Orbek-Key") {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		json.NewEncoder(w).Encode(s.e.view())
	})
	mux.HandleFunc("/changelog.md", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		io.WriteString(w, changelogMD)
	})
	mux.HandleFunc("/api/op", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		if !s.guard(w, r, "X-Orbek-Key") {
			return
		}
		raw, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		var req opReq
		if json.Unmarshal(raw, &req) != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.handleOp(req))
	})
	// the plugin inside Discord reports here and picks up what to do
	mux.HandleFunc("/bridge/sync", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || !s.guard(w, r, "X-Orbek-Key") {
			return
		}
		raw, _ := io.ReadAll(io.LimitReader(r.Body, 16<<20))
		var sy Sync
		if err := json.Unmarshal(raw, &sy); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.e.onSync(sy))
	})
	// a second start of the program asks the running one to show its panel
	mux.HandleFunc("/api/show", func(w http.ResponseWriter, r *http.Request) {
		if !s.guard(w, r, "X-Orbek-Key") {
			return
		}
		s.show()
		fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !localHost(r) { // the page carries the key: local addresses only
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" || name == "panel" {
			name = "index.html"
		}
		raw, err := uiFS.ReadFile("ui/" + name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		switch {
		case strings.HasSuffix(name, ".html"):
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			raw = []byte(strings.ReplaceAll(string(raw), "__KEY__", s.key))
		case strings.HasSuffix(name, ".json"):
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
		case strings.HasSuffix(name, ".svg"):
			w.Header().Set("Content-Type", "image/svg+xml")
		case strings.HasSuffix(name, ".ico"):
			w.Header().Set("Content-Type", "image/x-icon")
		}
		w.Write(raw)
	})
	return mux
}

// showMini opens the small window with just the quests' rewards and progress rings.
func (s *server) showMini() {
	if err := openMini(fmt.Sprintf("http://127.0.0.1:%d/panel?mini=1", s.port), s.e.dir); err != nil {
		s.e.log.Err("L.panelFail", err)
	}
}

// show opens the panel window; an open panel closes itself when it sees the new generation.
func (s *server) show() {
	s.e.mu.Lock()
	s.e.winGen++
	gen := s.e.winGen
	s.e.mu.Unlock()
	if err := openPanel(fmt.Sprintf("http://127.0.0.1:%d/panel?g=%d", s.port, gen), s.e.dir); err != nil {
		s.e.log.Err("L.panelFail", err)
	}
}

func (e *Engine) statusLine() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.connectedLocked() {
		return appName + " – " + tr("tray.noDiscord")
	}
	open := 0
	now := time.Now().UnixMilli()
	for i := range e.last.Quests {
		if e.last.Quests[i].open(now) {
			open++
		}
	}
	return fmt.Sprintf("%s – %s", appName, fmt.Sprintf(tr("tray.status"), open))
}
