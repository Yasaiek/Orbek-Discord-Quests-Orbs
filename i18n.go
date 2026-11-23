package main

import (
	"encoding/json"
	"os"
	"strings"
	"sync/atomic"
)

// The interface, the journal and the tray exist in these languages (ui/lang/<code>.json).
var langCodes = []string{"en", "pl", "de", "fr", "es", "it", "pt_BR", "ru", "uk", "tr", "ja", "zh_CN"}

var dicts = map[string]map[string]string{}
var curLang atomic.Value // string

func loadDicts() {
	for _, c := range langCodes {
		raw, err := uiFS.ReadFile("ui/lang/" + c + ".json")
		if err != nil {
			continue
		}
		m := map[string]string{}
		if json.Unmarshal(raw, &m) == nil {
			dicts[c] = m
		}
	}
	curLang.Store("en")
}

// tr returns the text for a key in the current language; an unknown key comes back unchanged,
// so plain (untranslated) messages pass through.
func tr(key string) string {
	lang, _ := curLang.Load().(string)
	if s, ok := dicts[lang][key]; ok {
		return s
	}
	if s, ok := dicts["en"][key]; ok {
		return s
	}
	return key
}

// matchLang maps a system locale name ("pl-PL", "pt_BR.UTF-8") to one of the supported languages.
func matchLang(tag string) string {
	t := strings.ToLower(strings.ReplaceAll(tag, "_", "-"))
	if i := strings.IndexAny(t, ".@"); i >= 0 {
		t = t[:i]
	}
	switch {
	case strings.HasPrefix(t, "pt"):
		return "pt_BR"
	case strings.HasPrefix(t, "zh"):
		return "zh_CN"
	}
	if i := strings.Index(t, "-"); i >= 0 {
		t = t[:i]
	}
	for _, c := range langCodes {
		if c == t {
			return c
		}
	}
	return "en"
}

// resolveLang: an explicit choice wins, otherwise the language of the system.
func resolveLang(pref string) string {
	for _, c := range langCodes {
		if c == pref {
			return c
		}
	}
	if l := os.Getenv("ORBEK_LANG"); l != "" {
		return matchLang(l)
	}
	return matchLang(systemLocale())
}

func (e *Engine) applyLang() {
	e.mu.Lock()
	pref := e.set.Lang
	e.mu.Unlock()
	curLang.Store(resolveLang(pref))
}
