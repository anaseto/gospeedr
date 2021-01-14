// +build !sdl

package main

import (
	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/drivers/tcell"
	tc "github.com/gdamore/tcell/v2"
)

var driver gruid.Driver

const TTF = false

func initDriver() {
	st := styler{}
	driver = tcell.NewDriver(tcell.Config{StyleManager: st})
}

// styler implements the tcell.StyleManager interface.
type styler struct{}

func (sty styler) GetStyle(st gruid.Style) tc.Style {
	ts := tc.StyleDefault
	switch st.Fg {
	case ColorYellow:
		ts = ts.Foreground(tc.ColorYellow)
	case ColorMagenta:
		ts = ts.Foreground(tc.ColorPurple)
	case ColorCyan:
		ts = ts.Foreground(tc.ColorTeal)
	case ColorGreen:
		ts = ts.Foreground(tc.ColorGreen)
	case ColorRed:
		ts = ts.Foreground(tc.ColorRed)
	}
	return ts
}
