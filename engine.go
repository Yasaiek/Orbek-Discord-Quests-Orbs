package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var version = "0.3.7"

const appName = "Orbek"

// Orbek never talks to Discord itself. Everything goes through the "Orbek" plugin inside the Discord app
// (Vencord): it sends the quest list here every couple of seconds and takes back what the user clicked.
// Requests to Discord therefore come from the real Discord client, and the account's token never leaves it.

type Settings struct {
	Lang        string    `json:"lang"`               // "" = the system's
	AutoMode    bool      `json:"autoMode"`           // the user allowed doing quests automatically
	AutoAll     bool      `json:"autoAll"`            // ...and open quests are done without asking
	AutoOnly    string    `json:"autoOnly,omitempty"` // up to 0.2.1: all | orbs | cosmetic (now AutoKinds)
	AutoKinds   []string  `json:"autoKinds"`          // which rewards the automat goes for: orbs, cosmetic, ingame, code, other
	AutoTasks   []string  `json:"autoTasks"`          // which tasks: video, play, activity, stream
	AutoGames   string    `json:"autoGames"`          // all | only (the games in GameList) | except (all but those)
	GameList    []string  `json:"gameList"`
	OpenOnStart bool      `json:"openOnStart"` // show the panel when the program starts
	Autostart   bool      `json:"autostart"`   // start with Windows
	Parallel    int       `json:"parallel"`    // how many quests at once (0 = no limit)
	AutoUpdate  bool      `json:"autoUpdate"`
	ClaimNote   ClaimNote `json:"claimNote"`  // how the last try of claiming went
	RefreshMin  int       `json:"refreshMin"` // how often the plugin reads the quest list again (minutes)
	AutoClaim   bool      `json:"autoClaim"`  // experimental: claim finished quests through Discord's own request
	// Discord activity: on, and what it shows (quest and progress, Orbs you have, Orbs to get, the buttons)
	RpcOn      bool              `json:"rpcOn"`
	RpcQuest   bool              `json:"rpcQuest"`
	RpcOrbs    bool              `json:"rpcOrbs"`
	RpcToGet   bool              `json:"rpcToGet"`
	RpcBtn     bool              `json:"rpcBtn"`
	RpcInit    bool              `json:"rpcInit"` // the choices above have their defaults
	OrbAnim    bool              `json:"orbAnim"` // the Orbs picture moves
	LookInit   bool              `json:"lookInit"`
	Stats      bool              `json:"stats"` // anonymous statistics for the website (on unless switched off)
	StatsInit  bool              `json:"statsInit"`
	VencordDir string            `json:"vencordDir"`         // where the Vencord source is (for installing the plugin)
	OrbTile    string            `json:"orbTile"`            // Discord's own Orbs pictures, as the plugin found them on Discord's quests page:
	OrbGlyph   string            `json:"orbGlyph"`           // the big reward tile and the small white sign
	OrbTiles   map[string]string `json:"orbTiles,omitempty"` // Orbs amount → Discord's tile for it (learnt by the plugin)
}

// Quest is what the plugin reports (already in a plain shape).
type Quest struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Game      string   `json:"game"`
	Publisher string   `json:"publisher"`
	Task      string   `json:"task"`
	Target    float64  `json:"target"`
	Progress  float64  `json:"progress"`
	Rewards   []Reward `json:"rewards"`
	Hero      string   `json:"hero"`
	Icon      string   `json:"icon"` // the game's tile
	Starts    int64    `json:"startsAt"`
	Expires   int64    `json:"expiresAt"`
	Enrolled  int64    `json:"enrolledAt"`
	Completed int64    `json:"completedAt"`
	Claimed   int64    `json:"claimedAt"`
	Auto      bool     `json:"auto"` // the plugin knows how to do this kind of task
}

type Reward struct {
	Type int     `json:"type"`
	Name string  `json:"name"`
	Orbs float64 `json:"orbs"`
	Img  string  `json:"img"`
}

const rCollectible, rOrbs = 3, 4

func (q *Quest) orbs() float64 {
	n := 0.0
	for _, r := range q.Rewards {
		if r.Type == rOrbs {
			n += r.Orbs
		}
	}
	return n
}

func (q *Quest) cosmetic() bool {
	for _, r := range q.Rewards {
		if r.Type == rCollectible {
			return true
		}
	}
	return false
}

func (q *Quest) open(now int64) bool {
	return q.Completed == 0 && (q.Expires == 0 || now < q.Expires) && (q.Starts == 0 || now >= q.Starts)
}

// Sync is one report of the plugin.
type Sync struct {
	V          int               `json:"v"`
	User       *BUser            `json:"user"`
	Quests     []Quest           `json:"quests"`
	Loaded     bool              `json:"loaded"` // the plugin has read the list at least once
	Err        string            `json:"err"`
	Orbs       *float64          `json:"orbs"`
	Runs       []RunState        `json:"runs"`
	Queue      []string          `json:"queue"`
	Events     []Event           `json:"events"`
	Desktop    bool              `json:"desktop"` // the Discord app (not a browser): games can be done
	Plugin     string            `json:"plugin"`  // the plugin's version
	OrbTile    string            `json:"orbTile"` // Discord's Orbs pictures (from Discord's own page), when the plugin saw them
	OrbGlyph   string            `json:"orbGlyph"`
	OrbTiles   map[string]string `json:"orbTiles,omitempty"`   // Orbs amount → the tile Discord shows with it
	ClaimDebug json.RawMessage   `json:"claimDebug,omitempty"` // how Discord claims (learnt from a claim by hand) and the last try
	OrbDebug   json.RawMessage   `json:"orbDebug,omitempty"`   // what the plugin saw around "Orbs" on Discord's page
}

type BUser struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type RunState struct {
	ID         string  `json:"id"`
	Progress   float64 `json:"progress"`
	ProgressAt int64   `json:"progressAt"` // when the progress last changed (the panel counts on from there)
	Note       string  `json:"note"`
	Since      int64   `json:"since"`
}

// Event: something the plugin did, for the journal.
type Event struct {
	K    string `json:"k"` // start | done | fail | stop | enrolled | enrollFail
	ID   string `json:"id"`
	Name string `json:"name"`
	Msg  string `json:"msg"`
}

// Cmd: something the user asked for, waiting to be picked up by the plugin.
type Cmd struct {
	Op  string   `json:"op"` // refresh | enroll | run | stop | stopAll | claim
	ID  string   `json:"id,omitempty"`
	IDs []string `json:"ids,omitempty"`
}

type Engine struct {
	dir string
	mu  sync.Mutex
	set Settings
	log *Logger

	last     Sync
	lastAt   time.Time // last report of the plugin
	cmds     []Cmd
	fails    map[string]time.Time
	autoSent map[string]time.Time // quests handed to the plugin by "do them by themselves"
	upd      updState
	setup    setupState

	quit chan struct{}

	games      map[string]int64 // games seen in quests (name → last seen)
	gameIcons  map[string]string
	rpcState   string // Discord activity: off | noid | idle | nodiscord | badid | refused | ok
	gamesSaved time.Time
	onQuit     func()
	winGen     int
	apiKey     string
	port       int
}

func newEngine(dir string) *Engine {
	e := &Engine{dir: dir, log: newLogger(dir), fails: map[string]time.Time{}, autoSent: map[string]time.Time{}, quit: make(chan struct{})}
	e.set = Settings{OpenOnStart: true, AutoUpdate: true}
	readJSON(filepath.Join(dir, "settings.json"), &e.set)
	if !e.set.RpcInit {
		e.set.RpcQuest, e.set.RpcOrbs, e.set.RpcToGet, e.set.RpcBtn, e.set.RpcInit = true, true, true, true, true
	}
	if e.set.ClaimNote.State == "testing" { // a test cut short by a restart
		e.set.ClaimNote = ClaimNote{State: "idle"}
	}
	if !e.set.StatsInit {
		e.set.Stats, e.set.StatsInit = true, true
	}
	if !e.set.LookInit {
		e.set.OrbAnim, e.set.LookInit = true, true
	}
	if e.set.RefreshMin <= 0 {
		e.set.RefreshMin = 10
	}
	e.set.migrateFilters()
	e.set.AutoOnly = ""
	e.loadGames()
	return e
}

func (e *Engine) saveSettings() { writeJSON(filepath.Join(e.dir, "settings.json"), e.set) }

// connectedLocked: the plugin reported recently.
func (e *Engine) connectedLocked() bool { return time.Since(e.lastAt) < 12*time.Second }

func (e *Engine) push(c Cmd) {
	e.mu.Lock()
	e.cmds = append(e.cmds, c)
	e.mu.Unlock()
}

// onSync takes a report of the plugin and returns what it should do.
func (e *Engine) onSync(s Sync) map[string]any {
	e.mu.Lock()
	defer e.mu.Unlock()
	if !e.connectedLocked() {
		name := ""
		if s.User != nil {
			name = s.User.Name
		}
		e.log.OK("L.connected", name)
	}
	e.last, e.lastAt = s, time.Now()
	if len(s.ClaimDebug) > 2 {
		os.WriteFile(filepath.Join(e.dir, "claim-debug.json"), s.ClaimDebug, 0o644)
	}
	for n, u := range s.OrbTiles {
		if strings.HasPrefix(u, "https://") && e.set.OrbTiles[n] != u {
			if e.set.OrbTiles == nil {
				e.set.OrbTiles = map[string]string{}
			}
			e.set.OrbTiles[n] = u
			e.saveSettings()
		}
	}
	if len(s.OrbDebug) > 2 {
		os.WriteFile(filepath.Join(e.dir, "orb-debug.json"), s.OrbDebug, 0o644)
	}
	e.noteGamesLocked(s.Quests)
	for _, f := range []struct{ got, have *string }{{&s.OrbTile, &e.set.OrbTile}, {&s.OrbGlyph, &e.set.OrbGlyph}} {
		if ic := *f.got; ic != "" && ic != *f.have && strings.HasPrefix(ic, "https://") {
			*f.have = ic
			e.saveSettings()
		}
	}
	for _, ev := range s.Events {
		switch ev.K {
		case "start":
			e.log.Info("L.runStart", ev.Name)
		case "done":
			e.log.OK("L.done", ev.Name)
		case "fail":
			e.fails[ev.ID] = time.Now()
			e.log.Warn("L.runFail", ev.Name, ev.Msg)
		case "stop":
			e.fails[ev.ID] = time.Now() // stopped by hand: "do them by themselves" must not restart it at once
			e.log.Info("L.runStopped", ev.Name)
		case "enrolled":
			e.log.OK("L.enrolled", ev.Name)
		case "enrollFail":
			e.log.Warn("L.enrollFail", ev.Name, ev.Msg)
		case "claimed":
			e.log.OK("L.claimed", ev.Name)
			e.claimResultLocked("ok", "")
		case "claimCaptcha":
			e.log.Warn("L.claimCaptcha", ev.Name)
			e.claimResultLocked("captcha", "")
		case "claimFail":
			e.log.Warn("L.claimFail", ev.Name, ev.Msg)
			e.claimResultLocked("failed", ev.Msg)
		case "claimNone":
			e.claimResultLocked("none", "")
		case "claimLearn":
			e.log.Warn("L.claimFail", ev.Name, ev.Msg)
			e.claimResultLocked("learn", ev.Msg)
		}
	}
	e.autoQueueLocked()
	cmds := e.cmds
	e.cmds = nil
	if cmds == nil {
		cmds = []Cmd{}
	}
	return map[string]any{"cmds": cmds, "auto": e.set.AutoMode, "parallel": e.set.Parallel, "claim": e.set.AutoClaim, "refresh": e.set.RefreshMin, "lang": curLang.Load()}
}

// desktopTask: done by the Discord app on the computer itself.
func desktopTask(t string) bool { return t == "PLAY_ON_DESKTOP" || t == "STREAM_ON_DESKTOP" }

// autoQueueLocked: with "do open quests by themselves" on, hands the plugin every quest it can do.
func (e *Engine) autoQueueLocked() {
	if !e.set.AutoMode || !e.set.AutoAll || len(e.last.Queue) > 0 || len(e.cmds) > 0 || !e.last.Loaded {
		return
	}
	if e.set.Parallel > 0 && len(e.last.Runs) >= e.set.Parallel {
		return
	}
	running := map[string]bool{}
	for _, r := range e.last.Runs {
		running[r.ID] = true
	}
	now := time.Now().UnixMilli()
	var qs []*Quest
	for i := range e.last.Quests {
		q := &e.last.Quests[i]
		if !q.open(now) || !q.Auto || running[q.ID] || (desktopTask(q.Task) && !e.last.Desktop) {
			continue
		}
		if at, ok := e.autoSent[q.ID]; ok && time.Since(at) < 2*time.Minute {
			continue
		}
		if at, ok := e.fails[q.ID]; ok && time.Since(at) < 3*time.Hour {
			continue
		}
		if !e.set.autoWants(q) {
			continue
		}
		qs = append(qs, q)
	}
	if len(qs) == 0 {
		return
	}
	sort.Slice(qs, func(i, j int) bool { return exp(qs[i]) < exp(qs[j]) })
	ids := make([]string, len(qs))
	for i, q := range qs {
		ids[i] = q.ID
		e.autoSent[q.ID] = time.Now()
	}
	e.cmds = append(e.cmds, Cmd{Op: "run", IDs: ids})
	e.log.Info("L.autoQueued", len(ids))
}

func exp(q *Quest) int64 {
	if q.Expires == 0 {
		return 1 << 62
	}
	return q.Expires
}

// run keeps the bridge file in place: it tells the plugin where to find this program.
func (e *Engine) run() {
	e.log.Info("L.started", version)
	e.writeBridge()
	go e.updateLoop()
	go usageLoop(e.quit)
	go e.discordLoop()
	go e.statsLoop("orbek")
	t := time.NewTicker(30 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-e.quit:
			return
		case <-t.C:
			e.writeBridge()
		}
	}
}

// bridgePath: a fixed place the plugin knows (%APPDATA%\Orbek\bridge.json).
func bridgePath() string {
	if p := os.Getenv("ORBEK_BRIDGE"); p != "" {
		return p
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(base, "Orbek", "bridge.json")
}

func (e *Engine) writeBridge() {
	p := bridgePath()
	if p == "" {
		return
	}
	os.MkdirAll(filepath.Dir(p), 0o700)
	exe, _ := os.Executable()
	e.mu.Lock()
	b := map[string]any{"port": e.port, "key": e.apiKey, "exe": exe, "pid": os.Getpid(), "version": version}
	e.mu.Unlock()
	raw, _ := json.Marshal(b)
	if old, _ := os.ReadFile(p); string(old) != string(raw) {
		os.WriteFile(p, raw, 0o600)
	}
	// where.json stays after the program ends: the plugin's "Open Orbek" starts it from there
	where, _ := json.Marshal(map[string]any{"exe": exe})
	if old, _ := os.ReadFile(filepath.Join(filepath.Dir(p), "where.json")); string(old) != string(where) {
		os.WriteFile(filepath.Join(filepath.Dir(p), "where.json"), where, 0o600)
	}
}

func (e *Engine) shutdown() {
	select {
	case <-e.quit:
	default:
		close(e.quit)
	}
	if p := bridgePath(); p != "" {
		var b struct {
			Pid int `json:"pid"`
		}
		if readJSON(p, &b) == nil && b.Pid == os.Getpid() {
			os.Remove(p)
		}
	}
}
