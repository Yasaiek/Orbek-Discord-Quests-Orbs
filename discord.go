package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Discord Rich Presence. While Orbek is doing a quest it tells the Discord client on this computer, and
// Discord shows it on the profile: the quest, its progress, the Orbs. Nothing is shown while Orbek only
// waits, and nothing at all unless the user switches it on. It talks to the client over its local pipe.

// discordAppID is Orbek's application in the Discord Developer Portal (names the activity, holds its pictures).
const discordAppID = "1574392419363848326"

// The pictures uploaded to that application under Rich Presence > Art Assets.
const (
	rpcLogo = "logo"
)

const repoURL = "https://github.com/Yasaiek/Orbek-Discord-Quests-Orbs"

type rpcActivity struct {
	Type       int         `json:"type"` // 0 = "Playing"
	Details    string      `json:"details,omitempty"`
	State      string      `json:"state,omitempty"`
	Timestamps *rpcTimes   `json:"timestamps,omitempty"`
	Assets     *rpcAssets  `json:"assets,omitempty"`
	Buttons    []rpcButton `json:"buttons,omitempty"`
	Instance   bool        `json:"instance"`
}

type rpcTimes struct {
	Start int64 `json:"start"`
}

type rpcAssets struct {
	LargeImage string `json:"large_image,omitempty"`
	LargeText  string `json:"large_text,omitempty"`
	SmallImage string `json:"small_image,omitempty"`
	SmallText  string `json:"small_text,omitempty"`
}

type rpcButton struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// rpcText fits a text into what Discord accepts for one field: 2 to 128 characters.
func rpcText(s string) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) > 128 {
		r := []rune(s)
		s = strings.TrimSpace(string(r[:127])) + "…"
	}
	if utf8.RuneCountInString(s) < 2 {
		return ""
	}
	return s
}

func firstNonEmpty(a ...string) string {
	for _, s := range a {
		if s != "" {
			return s
		}
	}
	return ""
}

// rpcCfg is what the user lets Discord show.
type rpcCfg struct{ On, Quest, Orbs, ToGet, Button bool }

// presence builds what Discord should show right now; nil when there is nothing to show.
func (e *Engine) presence(now time.Time) (cfg rpcCfg, id string, act *rpcActivity) {
	e.mu.Lock()
	defer e.mu.Unlock()
	cfg = rpcCfg{e.set.RpcOn, e.set.RpcQuest, e.set.RpcOrbs, e.set.RpcToGet, e.set.RpcBtn}
	id = discordAppID
	if now.Before(e.rpcTestUntil) {
		return cfg, id, rpcTestActivity(e.rpcTestUntil)
	}
	if !cfg.On || id == "" || !e.connectedLocked() || len(e.last.Runs) == 0 {
		return cfg, id, nil // only while a quest is being done
	}
	run := e.last.Runs[0]
	var q *Quest
	for i := range e.last.Quests {
		if e.last.Quests[i].ID == run.ID {
			q = &e.last.Quests[i]
		}
	}
	if q == nil {
		return cfg, id, nil
	}
	act = &rpcActivity{Type: 0, Assets: &rpcAssets{LargeImage: rpcLogo, SmallImage: taskKind(q.Task)}}
	act.Timestamps = &rpcTimes{Start: firstPositive(run.Since, now.UnixMilli()) / 1000}
	if cfg.Quest {
		act.Details = fmt.Sprintf(tr("rpc.doing"), q.Name)
		if q.Target > 0 {
			p := min(run.Progress, q.Target)
			act.State = fmt.Sprintf(tr("rpc.prog"), mins(p), mins(q.Target), int(p*100/q.Target))
		}
		if ic := q.Icon; strings.HasPrefix(ic, "https://") && len(ic) <= 256 {
			act.Assets.LargeImage = ic
		}
		act.Assets.LargeText = firstNonEmpty(q.Game, q.Name)
	} else {
		act.Details = tr("rpc.doingAny")
	}
	if n := len(e.last.Runs); n > 1 {
		more := fmt.Sprintf(tr("rpc.more"), n-1)
		if act.State == "" {
			act.State = more
		} else {
			act.State += " · " + more
		}
	}
	var extra []string
	if cfg.Orbs && e.last.Orbs != nil {
		extra = append(extra, fmt.Sprintf(tr("rpc.orbs"), int(*e.last.Orbs)))
	}
	if cfg.ToGet {
		left, nowMs := 0.0, now.UnixMilli()
		for i := range e.last.Quests {
			if e.last.Quests[i].open(nowMs) {
				left += e.last.Quests[i].orbs()
			}
		}
		if left > 0 {
			extra = append(extra, fmt.Sprintf(tr("rpc.toGet"), int(left)))
		}
	}
	act.Assets.SmallText = firstNonEmpty(strings.Join(extra, " · "), tr("f.t."+taskKind(q.Task)))
	if act.Assets.LargeText == "" {
		act.Assets.LargeText = appName + " " + version
	}
	act.Details, act.State = rpcText(act.Details), rpcText(act.State)
	act.Assets.LargeText, act.Assets.SmallText = rpcText(act.Assets.LargeText), rpcText(act.Assets.SmallText)
	if cfg.Button {
		act.Buttons = []rpcButton{{Label: rpcText(tr("rpc.btn")), URL: siteURL}, {Label: "GitHub", URL: repoURL}}
	}
	return cfg, id, act
}

func firstPositive(a ...int64) int64 {
	for _, v := range a {
		if v > 0 {
			return v
		}
	}
	return 0
}

// mins: 900 seconds → "15 min", 59 → "59 s".
func mins(sec float64) string {
	if sec < 60 {
		return strconv.Itoa(int(sec)) + " s"
	}
	return strconv.Itoa(int(sec)/60) + " min"
}

// --- the connection ---

type discordRPC struct {
	conn    net.Conn
	appID   string
	last    string // the activity Discord has, as JSON; "" = none
	sentAt  time.Time
	retryAt time.Time
	nonce   int
	err     string // the last error, shown in the settings
}

var errDiscordID = errors.New("discord does not know this application id")

func (d *discordRPC) close() {
	if d.conn != nil {
		d.conn.Close()
	}
	d.conn, d.last = nil, ""
}

func (d *discordRPC) write(op uint32, v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	buf := make([]byte, 8+len(raw))
	binary.LittleEndian.PutUint32(buf, op)
	binary.LittleEndian.PutUint32(buf[4:], uint32(len(raw)))
	copy(buf[8:], raw)
	_, err = d.conn.Write(buf)
	return err
}

func (d *discordRPC) read() (uint32, map[string]any, error) {
	var head [8]byte
	if _, err := io.ReadFull(d.conn, head[:]); err != nil {
		return 0, nil, err
	}
	n := binary.LittleEndian.Uint32(head[4:])
	if n > 1<<20 {
		return 0, nil, errors.New("discord: frame too long")
	}
	raw := make([]byte, n)
	if _, err := io.ReadFull(d.conn, raw); err != nil {
		return 0, nil, err
	}
	var m map[string]any
	json.Unmarshal(raw, &m)
	return binary.LittleEndian.Uint32(head[:]), m, nil
}

// connect finds the running Discord client (it listens on one of ten pipes) and introduces the application.
func (d *discordRPC) connect(appID string) error {
	var err error
	for i := 0; i < 10; i++ {
		var c net.Conn
		if c, err = dialDiscord(i); err != nil {
			continue
		}
		d.conn, d.appID, d.last = c, appID, ""
		c.SetDeadline(time.Now().Add(5 * time.Second))
		if err = d.write(0, map[string]any{"v": 1, "client_id": appID}); err == nil {
			var op uint32
			var m map[string]any
			if op, m, err = d.read(); err == nil {
				if op == 1 && str(m["evt"]) == "READY" {
					return nil
				}
				err = fmt.Errorf("discord: handshake refused (%d %s)", num(m["code"]), str(m["message"]))
				if num(m["code"]) == 4000 {
					err = errDiscordID
				}
			}
		}
		d.close()
		if err == errDiscordID {
			return err
		}
	}
	if err == nil {
		err = errors.New("discord is not running")
	}
	return err
}

// set shows the activity (nil clears it). A refusal of the activity itself comes back as rpcRefused.
type rpcRefused string

func (r rpcRefused) Error() string { return "discord refused the activity: " + string(r) }

func (d *discordRPC) set(act *rpcActivity) error {
	d.nonce++
	nonce := strconv.Itoa(d.nonce)
	args := map[string]any{"pid": os.Getpid()}
	if act != nil {
		args["activity"] = act
	}
	d.conn.SetDeadline(time.Now().Add(5 * time.Second))
	if err := d.write(1, map[string]any{"cmd": "SET_ACTIVITY", "args": args, "nonce": nonce}); err != nil {
		return err
	}
	for {
		op, m, err := d.read()
		if err != nil {
			return err
		}
		switch {
		case op == 3: // ping
			if err := d.write(4, m); err != nil {
				return err
			}
		case op == 2:
			return fmt.Errorf("discord closed the connection (%d %s)", num(m["code"]), str(m["message"]))
		case op == 1 && str(m["nonce"]) == nonce:
			if str(m["evt"]) == "ERROR" {
				return rpcRefused(str(dig(m, "data", "message")))
			}
			return nil
		}
	}
}

// discordLoop keeps Discord in step with what Orbek is doing.
func (e *Engine) discordLoop() {
	var d discordRPC
	t := time.NewTicker(3 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-e.quit:
			if d.conn != nil {
				d.set(nil)
				d.close()
			}
			return
		case <-t.C:
		}
		state := e.discordStep(&d, time.Now())
		e.mu.Lock()
		e.rpcState, e.rpcErr = state, d.err
		e.mu.Unlock()
	}
}

// discordStep returns the state shown in the settings: off | noid | nodiscord | badid | refused | ok.
func (e *Engine) discordStep(d *discordRPC, now time.Time) string {
	cfg, id, act := e.presence(now)
	if act == nil {
		if d.conn != nil {
			d.set(nil)
			d.close()
		}
		d.retryAt = time.Time{}
		switch {
		case !cfg.On:
			return "off"
		case id == "":
			return "noid"
		}
		return "idle"
	}
	if d.conn != nil && d.appID != id {
		d.set(nil)
		d.close()
		d.retryAt = time.Time{}
	}
	if d.conn == nil {
		if now.Before(d.retryAt) {
			return e.rpcStateNow()
		}
		if err := d.connect(id); err != nil {
			d.err = err.Error()
			d.retryAt = now.Add(20 * time.Second)
			if err == errDiscordID {
				e.log.Debug("discord: %v", err)
				return "badid"
			}
			return "nodiscord"
		}
		e.log.Debug("discord: connected")
		d.sentAt = time.Time{}
	}
	raw, _ := json.Marshal(act)
	if string(raw) == d.last || now.Sub(d.sentAt) < 4*time.Second { // Discord takes five changes in twenty seconds
		if d.last == "" {
			return e.rpcStateNow()
		}
		return "ok"
	}
	err := d.set(act)
	d.sentAt = now
	var refused rpcRefused
	switch {
	case errors.As(err, &refused):
		d.err = err.Error()
		e.log.Debug("discord: %v", err)
		d.last = string(raw) // sending the same thing again would be refused again
		return "refused"
	case err != nil:
		d.err = err.Error()
		e.log.Debug("discord: %v", err)
		d.close()
		d.retryAt = now.Add(10 * time.Second)
		return "nodiscord"
	}
	d.last = string(raw)
	d.err = ""
	return "ok"
}

func (e *Engine) rpcStateNow() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.rpcState
}

// rpcTestActivity: what "Test" in the settings shows for 30 seconds, whatever the program is doing.
func rpcTestActivity(until time.Time) *rpcActivity {
	return &rpcActivity{Type: 0, Details: rpcText(tr("rpc.testD")), State: rpcText(tr("rpc.testS")),
		Timestamps: &rpcTimes{Start: until.Add(-30 * time.Second).Unix()},
		Assets:     &rpcAssets{LargeImage: rpcLogo, LargeText: rpcText(appName + " " + version)},
		Buttons:    []rpcButton{{Label: rpcText(tr("rpc.btn")), URL: siteURL}, {Label: "GitHub", URL: repoURL}}}
}
