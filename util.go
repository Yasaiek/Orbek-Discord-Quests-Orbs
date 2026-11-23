package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// dig walks nested JSON (map[string]any / []any) by keys; returns nil when anything is missing.
func dig(v any, path ...string) any {
	for _, k := range path {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = m[k]
	}
	return v
}

func str(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return fmt.Sprintf("%.0f", t)
	case json.Number:
		return t.String()
	}
	return ""
}

func num(v any) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case string:
		var n int64
		fmt.Sscan(t, &n)
		return n
	}
	return 0
}

func boolean(v any) bool { b, _ := v.(bool); return b }

func list(v any) []any { l, _ := v.([]any); return l }

func parseTime(v any) time.Time {
	t, err := time.Parse(time.RFC3339, str(v))
	if err != nil {
		return time.Time{}
	}
	return t
}

func randHex(n int) string {
	b := make([]byte, n/2+1)
	rand.Read(b)
	return hex.EncodeToString(b)[:n]
}

// writeJSON saves atomically: a crash mid-write never leaves a broken file.
func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// ---- log: file + in-memory tail for the panel ----

type LogLine struct {
	T int64  `json:"t"`
	L string `json:"l"` // info | ok | warn | err
	M string `json:"m"`
}

type Logger struct {
	mu    sync.Mutex
	lines []LogLine
	file  string
}

func newLogger(dir string) *Logger {
	l := &Logger{file: filepath.Join(dir, "log.txt")}
	if st, err := os.Stat(l.file); err == nil && st.Size() > 2<<20 { // keep the file small
		os.Rename(l.file, filepath.Join(dir, "log.old.txt"))
	}
	return l
}

func (l *Logger) add(level, format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, LogLine{T: now.UnixMilli(), L: level, M: msg})
	if len(l.lines) > 300 {
		l.lines = l.lines[len(l.lines)-300:]
	}
	if f, err := os.OpenFile(l.file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600); err == nil {
		fmt.Fprintf(f, "%s %-4s %s\n", now.Format("2006-01-02 15:04:05"), strings.ToUpper(level), msg)
		f.Close()
	}
}

// The first argument of these is a translation key (see i18n.go); a plain text works too.
func (l *Logger) Info(key string, a ...any) { l.add("info", tr(key), a...) }
func (l *Logger) OK(key string, a ...any)   { l.add("ok", tr(key), a...) }
func (l *Logger) Warn(key string, a ...any) { l.add("warn", tr(key), a...) }
func (l *Logger) Err(key string, a ...any)  { l.add("err", tr(key), a...) }

// Debug goes to the file only (details useful when something has to be diagnosed).
func (l *Logger) Debug(format string, a ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if f, err := os.OpenFile(l.file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600); err == nil {
		fmt.Fprintf(f, "%s DBG  %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, a...))
		f.Close()
	}
}

func (l *Logger) tail() []LogLine {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]LogLine, len(l.lines))
	copy(out, l.lines)
	return out
}

func (l *Logger) clear() {
	l.mu.Lock()
	l.lines = nil
	l.mu.Unlock()
}
