package main

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Where the program keeps its files. By default %LOCALAPPDATA%\Yasaiek\<program>, a folder shared by
// Yasaiek's programs, each in its own subfolder. In the settings another place can be chosen; the
// choice is kept outside the data itself, in %LOCALAPPDATA%\Yasaiek\<program>-location.json.

type location struct {
	Root string `json:"root,omitempty"` // chosen place (the program's folder is made inside it); empty: the default one
	From string `json:"from,omitempty"` // the files are still there: move them on the next start
}

func defaultRoot() string {
	if d := os.Getenv("LOCALAPPDATA"); d != "" {
		return filepath.Join(d, "Yasaiek")
	}
	if d, err := os.UserConfigDir(); err == nil {
		return filepath.Join(d, "Yasaiek")
	}
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".yasaiek")
}

func locationPath() string { return filepath.Join(defaultRoot(), appName+"-location.json") }

func readLocation() (l location) {
	readJSON(locationPath(), &l)
	return
}

func saveLocation(l location) error {
	if l.Root == "" && l.From == "" {
		if err := os.Remove(locationPath()); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	os.MkdirAll(defaultRoot(), 0o700)
	return writeJSON(locationPath(), l)
}

// dataRoot: the place chosen in the settings, or the default one.
func dataRoot() string {
	if r := readLocation().Root; r != "" {
		return r
	}
	return defaultRoot()
}

func dataDirIn(root string) string { return filepath.Join(root, appName) }

func sameDir(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

func inside(dir, parent string) bool {
	rel, err := filepath.Rel(parent, dir)
	return err == nil && rel != "." && !strings.HasPrefix(rel, "..")
}

// settleData moves the files where they belong: after the folder was changed in the settings, or from
// where older versions kept them (legacy). Called once the previous copy of the program has let go of them.
func settleData(dir string, legacy func(dir string)) {
	if l := readLocation(); l.From != "" {
		if !sameDir(l.From, dir) && !hasData(dir) {
			moveTree(l.From, dir)
		}
		l.From = ""
		saveLocation(l)
	}
	if !hasData(dir) && legacy != nil {
		legacy(dir)
	}
	os.MkdirAll(dir, 0o700)
}

func emptyDir(dir string) bool {
	ents, err := os.ReadDir(dir)
	return err == nil && len(ents) == 0
}

// moveTree moves a folder: a rename when it can, otherwise a copy and removing the old one. The browser
// profile of an open panel window is in use and is left behind (a new one is made).
func moveTree(src, dst string) error {
	if st, err := os.Stat(src); err != nil || !st.IsDir() {
		return errors.New("nothing to move")
	}
	os.MkdirAll(filepath.Dir(dst), 0o700)
	if emptyDir(dst) {
		os.Remove(dst)
	}
	if os.Rename(src, dst) == nil {
		return nil
	}
	if err := copyTree(src, dst); err != nil {
		return err
	}
	os.RemoveAll(src)
	return nil
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() && rel == "panel-profile" {
			return filepath.SkipDir
		}
		t := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(t, 0o700)
		}
		if d.Type().IsRegular() {
			copyFile(p, t)
		}
		return nil
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	return err
}

// chooseRoot checks a place picked in the settings and notes it, so that the next start moves the files.
// An empty root means the default place.
func chooseRoot(root, cur string) (string, error) {
	if root != "" {
		root = filepath.Clean(strings.TrimSpace(strings.Trim(root, `"`)))
		if !filepath.IsAbs(root) {
			return "", errors.New("bad")
		}
	}
	dir := dataDirIn(root)
	if root == "" || sameDir(root, defaultRoot()) {
		root, dir = "", dataDirIn(defaultRoot())
	}
	if sameDir(dir, cur) {
		return "", errors.New("same")
	}
	if inside(dir, cur) {
		return "", errors.New("bad")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", errors.New("bad")
	}
	test := filepath.Join(dir, ".write-test")
	if err := os.WriteFile(test, []byte("ok"), 0o600); err != nil {
		return "", errors.New("bad")
	}
	os.Remove(test)
	if err := saveLocation(location{Root: root, From: cur}); err != nil {
		return "", errors.New("bad")
	}
	return dir, nil
}
