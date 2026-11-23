//go:build windows

package main

import (
	"embed"
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"fyne.io/systray"
	"golang.org/x/sys/windows/registry"
)

//go:embed assets/icon.ico
var iconICO []byte

//go:embed assets/menu
var menuIcons embed.FS

func menuIcon(name string) []byte {
	b, _ := menuIcons.ReadFile("assets/menu/" + name + ".ico")
	return b
}

// runTray puts the icon next to the clock and blocks until the user quits. Right click opens Orbek's menu:
// state, how much of the PC it takes, the windows, stopping, updates, the website and quitting.
func runTray(e *Engine, a trayActions) {
	darkMenus()
	systray.Run(func() {
		systray.SetIcon(iconICO)
		systray.SetTitle(appName)
		systray.SetTooltip(appName)
		systray.SetOnTapped(a.Open)

		head := systray.AddMenuItem(appName+" "+version, "")
		head.SetIcon(iconICO)
		mState := systray.AddMenuItem("", "")
		mUse := systray.AddMenuItem(usageLine(), "")
		mUse.SetIcon(menuIcon("usage"))
		systray.AddSeparator()
		mOpen := systray.AddMenuItem(tr("tray.show"), "")
		mOpen.SetIcon(menuIcon("open"))
		mMini := systray.AddMenuItem(tr("tray.mini"), "")
		mMini.SetIcon(menuIcon("mini"))
		mStop := systray.AddMenuItem(tr("tray.stopAll"), "")
		mStop.SetIcon(menuIcon("stop"))
		systray.AddSeparator()
		mUpd := systray.AddMenuItem(tr("tray.update"), "")
		mUpd.SetIcon(menuIcon("update"))
		mSite := systray.AddMenuItem(tr("tray.site"), "")
		mSite.SetIcon(menuIcon("site"))
		mDir := systray.AddMenuItem(tr("tray.folder"), "")
		mDir.SetIcon(menuIcon("folder"))
		systray.AddSeparator()
		mQuit := systray.AddMenuItem(tr("quit"), "")
		mQuit.SetIcon(menuIcon("quit"))

		lastDot := ""
		refresh := func() {
			st := e.trayState()
			line := st.Line
			systray.SetTooltip(appName + " – " + line + "\n" + usageLine())
			mState.SetTitle(line)
			if dot := map[bool]string{true: "on", false: "off"}[st.Connected]; dot != lastDot {
				mState.SetIcon(menuIcon(dot))
				lastDot = dot
			}
			mUse.SetTitle(usageLine())
			if st.Running > 0 {
				mStop.SetTitle(fmt.Sprintf(tr("tray.stopN"), st.Running))
				mStop.Enable()
			} else {
				mStop.SetTitle(tr("tray.stopAll"))
				mStop.Disable()
			}
			mOpen.SetTitle(tr("tray.show"))
			mMini.SetTitle(tr("tray.mini"))
			mUpd.SetTitle(tr("tray.update"))
			mSite.SetTitle(tr("tray.site"))
			mDir.SetTitle(tr("tray.folder"))
			mQuit.SetTitle(tr("quit"))
		}
		refresh()
		go func() {
			t := time.NewTicker(2 * time.Second)
			defer t.Stop()
			for {
				select {
				case <-mOpen.ClickedCh:
					a.Open()
				case <-head.ClickedCh:
					a.Open()
				case <-mState.ClickedCh:
					a.Open()
				case <-mUse.ClickedCh:
					a.Open()
				case <-mMini.ClickedCh:
					a.Mini()
				case <-mStop.ClickedCh:
					a.StopAll()
				case <-mUpd.ClickedCh:
					a.Update()
				case <-mSite.ClickedCh:
					openURL(siteURL)
				case <-mDir.ClickedCh:
					if exe, err := os.Executable(); err == nil {
						exec.Command("explorer", filepath.Dir(exe)).Start()
					}
				case <-mQuit.ClickedCh:
					systray.Quit()
					return
				case <-e.quit:
					systray.Quit()
					return
				case <-t.C:
					refresh()
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

// ---- own taskbar button ----
//
// The panel is an app window of Edge/Chrome. Windows groups the windows of one program under one taskbar
// button, so Orbek's window landed under Loorek's (also a browser app window). Giving Orbek's windows their
// own AppUserModelID makes them a taskbar button of their own, with Orbek's icon and name.

const orbekAppID = "Yasaiek.Orbek"

type winGUID struct {
	d1     uint32
	d2, d3 uint16
	d4     [8]byte
}
type propKey struct {
	fmtid winGUID
	pid   uint32
}
type propVariant struct {
	vt         uint16
	r1, r2, r3 uint16
	val        uintptr
	pad        uintptr
}

var (
	user32             = syscall.NewLazyDLL("user32.dll")
	procEnumWindows    = user32.NewProc("EnumWindows")
	procGetWindowText  = user32.NewProc("GetWindowTextW")
	procGetClassName   = user32.NewProc("GetClassNameW")
	procIsVisible      = user32.NewProc("IsWindowVisible")
	procPropStoreForWn = syscall.NewLazyDLL("shell32.dll").NewProc("SHGetPropertyStoreForWindow")
	procCoInit         = syscall.NewLazyDLL("ole32.dll").NewProc("CoInitializeEx")

	iidPropertyStore = winGUID{0x886D8EEB, 0x8CF2, 0x4446, [8]byte{0x8D, 0x02, 0xCD, 0xBA, 0x1D, 0xBD, 0xCF, 0x99}}
	appModelFmt      = winGUID{0x9F4C2855, 0x9F79, 0x4B39, [8]byte{0xA8, 0xD0, 0xE1, 0xD4, 0x2D, 0xE1, 0xD5, 0xF3}}

	enumFound []uintptr
	enumCB    = syscall.NewCallback(func(h, _ uintptr) uintptr {
		if v, _, _ := procIsVisible.Call(h); v == 0 {
			return 1
		}
		buf := make([]uint16, 64)
		procGetClassName.Call(h, uintptr(unsafe.Pointer(&buf[0])), 64)
		if syscall.UTF16ToString(buf) != "Chrome_WidgetWin_1" {
			return 1
		}
		procGetWindowText.Call(h, uintptr(unsafe.Pointer(&buf[0])), 64)
		if syscall.UTF16ToString(buf) == appName { // an app window shows just the page's title; a normal tab adds " - Microsoft Edge"
			enumFound = append(enumFound, h)
		}
		return 1
	})
)

// tagWindows watches for Orbek's browser windows for a while and gives each its own taskbar identity.
func tagWindows() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	procCoInit.Call(0, 2) // COINIT_APARTMENTTHREADED
	exe, _ := os.Executable()
	done := map[uintptr]bool{}
	for end := time.Now().Add(25 * time.Second); time.Now().Before(end); time.Sleep(400 * time.Millisecond) {
		enumFound = enumFound[:0]
		procEnumWindows.Call(enumCB, 0)
		for _, h := range enumFound {
			if !done[h] {
				done[h] = setWindowAppID(h, exe)
			}
		}
	}
}

func setWindowAppID(hwnd uintptr, exe string) bool {
	var obj *struct{ vt *[8]uintptr } // an IPropertyStore
	if hr, _, _ := procPropStoreForWn.Call(hwnd, uintptr(unsafe.Pointer(&iidPropertyStore)), uintptr(unsafe.Pointer(&obj))); hr != 0 || obj == nil {
		return false
	}
	vt, ps := obj.vt, uintptr(unsafe.Pointer(obj))
	defer syscall.SyscallN(vt[2], ps) // Release
	set := func(pid uint32, s string) {
		p, err := syscall.UTF16PtrFromString(s)
		if err != nil {
			return
		}
		k := propKey{appModelFmt, pid}
		v := propVariant{vt: 31, val: uintptr(unsafe.Pointer(p))}                             // VT_LPWSTR
		syscall.SyscallN(vt[6], ps, uintptr(unsafe.Pointer(&k)), uintptr(unsafe.Pointer(&v))) // SetValue
		runtime.KeepAlive(p)
	}
	if exe != "" {
		set(2, `"`+exe+`"`) // RelaunchCommand (a pinned button starts Orbek)
		set(4, appName)     // RelaunchDisplayNameResource
		set(3, exe+",0")    // RelaunchIconResource
	}
	set(5, orbekAppID)          // ID
	syscall.SyscallN(vt[7], ps) // Commit
	return true
}

// darkMenus makes the program's menus (the tray menu) dark, whatever the Windows theme. It uses the same
// switch Windows' own apps use (uxtheme ordinals 135 SetPreferredAppMode and 136 FlushMenuThemes, Windows 10
// 1903 and newer); on older Windows the call does nothing harmful.
func darkMenus() {
	h, err := windows.LoadLibrary("uxtheme.dll")
	if err != nil {
		return
	}
	if p, err := windows.GetProcAddressByOrdinal(h, 135); err == nil {
		syscall.SyscallN(p, 2) // ForceDark
	}
	if p, err := windows.GetProcAddressByOrdinal(h, 136); err == nil {
		syscall.SyscallN(p)
	}
}

// notify shows a Windows notification from the tray icon, with the system sound.
func notify(title, text string) { systray.Notify(title, text, false) }
