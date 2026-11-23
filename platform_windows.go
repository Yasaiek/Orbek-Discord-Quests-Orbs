//go:build windows

package main

import (
	_ "embed"
	"os"
	"os/exec"
	"syscall"
	"time"
	"unsafe"

	"fyne.io/systray"
	"golang.org/x/sys/windows/registry"
)

//go:embed assets/icon.ico
var iconICO []byte

// runTray puts the icon next to the clock and blocks until the user quits.
func runTray(e *Engine, open func()) {
	systray.Run(func() {
		systray.SetIcon(iconICO)
		systray.SetTitle(appName)
		systray.SetTooltip(appName)
		systray.SetOnTapped(open)
		mOpen := systray.AddMenuItem(tr("tray.show"), "")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem(tr("quit"), "")
		go func() {
			t := time.NewTicker(5 * time.Second)
			defer t.Stop()
			for {
				select {
				case <-mOpen.ClickedCh:
					open()
				case <-mQuit.ClickedCh:
					systray.Quit()
					return
				case <-e.quit:
					systray.Quit()
					return
				case <-t.C:
					systray.SetTooltip(e.statusLine())
					mOpen.SetTitle(tr("tray.show"))
					mQuit.SetTitle(tr("quit"))
				}
			}
		}()
	}, func() {})
}

// systemLocale is the user's Windows display language, like "pl-PL".
func systemLocale() string {
	buf := make([]uint16, 85)
	r, _, _ := syscall.NewLazyDLL("kernel32.dll").NewProc("GetUserDefaultLocaleName").Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

const runKey = `Software\Microsoft\Windows\CurrentVersion\Run`

// setAutostart adds or removes this program from the user's "start with Windows" list.
func setAutostart(on bool) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if !on {
		err := k.DeleteValue(appName)
		if err == registry.ErrNotExist {
			return nil
		}
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return k.SetStringValue(appName, `"`+exe+`" --tray`)
}

// hideWindow keeps the console programs the setup runs from flashing a window.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000} // CREATE_NO_WINDOW
}
