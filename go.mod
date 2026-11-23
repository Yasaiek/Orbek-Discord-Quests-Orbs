module orbek

go 1.24.1

toolchain go1.24.7

require (
	fyne.io/systray v0.0.0
	golang.org/x/sys v0.31.0
)

require github.com/godbus/dbus/v5 v5.1.0 // indirect

replace fyne.io/systray => ./third_party/systray

// golang.org/x/* fetched from their GitHub mirrors
replace golang.org/x/sys => github.com/golang/sys v0.31.0
