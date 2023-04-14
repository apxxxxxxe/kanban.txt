package tui

import "github.com/rivo/tview"

type InputBox struct {
	*tview.InputField
	Mode rune
}
