package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const defaultPort = 47626

type instanceFile struct {
	Port int    `json:"port"`
	Key  string `json:"key"`
}

func main() {
	trayOnly, updated := false, false
	for _, a := range os.Args[1:] {
		switch a {
		case "--tray":
			trayOnly = true
		case "--updated":
			updated = true
		}
	}
	loadDicts()
	dir := os.Getenv("ORBEK_DATA") // everything the program keeps lives in "data" next to the executable
	if dir == "" {
		exe, err := os.Executable()
		if err != nil {
			exe = os.Args[0]
		}
		dir = filepath.Join(filepath.Dir(exe), "data")
	}
	os.MkdirAll(dir, 0o700)
	instPath := filepath.Join(dir, "instance.json")

	// already running? ask that copy to show its panel and leave (not when this start is an update taking over)
	var inst instanceFile
	if !updated && readJSON(instPath, &inst) == nil && inst.Port > 0 {
		req, _ := http.NewRequest("GET", "http://127.0.0.1:"+strconv.Itoa(inst.Port)+"/api/show", nil)
		req.Header.Set("X-Orbek-Key", inst.Key)
		if res, err := (&http.Client{Timeout: 2 * time.Second}).Do(req); err == nil {
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
			if res.StatusCode == 200 {
				return
			}
		}
	}

	// An update comes up on the same port and with the same key, so an open panel simply carries on.
	ports := []int{}
	if p, _ := strconv.Atoi(os.Getenv("ORBEK_PORT")); p > 0 {
		ports = append(ports, p)
	}
	ports = append(ports, defaultPort)
	var ln net.Listener
	var err error
	for n, p := range ports {
		ln, err = net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(p))
		for i := 0; err != nil && n == 0 && updated && i < 40; i++ { // the previous version is still letting go
			time.Sleep(200 * time.Millisecond)
			ln, err = net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(p))
		}
		if err == nil {
			break
		}
	}
	if err != nil {
		ln, err = net.Listen("tcp", "127.0.0.1:0")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot open a local port:", err)
		os.Exit(1)
	}
	e := newEngine(dir)
	srv := &server{e: e, key: randHex(32), port: ln.Addr().(*net.TCPAddr).Port}
	if k := os.Getenv("ORBEK_KEY"); k != "" {
		srv.key = k
	}
	if g, err := strconv.Atoi(os.Getenv("ORBEK_GEN")); err == nil {
		e.winGen = g
	}
	writeJSON(instPath, instanceFile{Port: srv.port, Key: srv.key})
	e.apiKey, e.port = srv.key, srv.port
	e.applyLang()
	cleanupUpdate()
	if updated {
		e.upd.Done = time.Now()
		e.log.OK("L.updDone", version)
	}
	e.onQuit = func() {
		os.Remove(instPath)
		e.shutdown()
		time.AfterFunc(300*time.Millisecond, func() { os.Exit(0) })
	}
	go http.Serve(ln, srv.routes())
	go e.run()

	e.mu.Lock()
	open := !trayOnly && e.set.OpenOnStart
	e.mu.Unlock()
	if open && os.Getenv("ORBEK_NO_WINDOW") == "" {
		srv.show()
	}
	runTray(e, trayActions{
		Open:    srv.show,
		Mini:    func() { go srv.showMini() },
		StopAll: func() { e.push(Cmd{Op: "stopAll"}) },
		Update:  func() { go e.updateCheck(true) },
	})
	os.Remove(instPath)
	e.shutdown()
	time.Sleep(300 * time.Millisecond)
}
