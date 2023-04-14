package tui

import (
	"github.com/rivo/tview"
)

func newTable(title string) *tview.Table {
	t := tview.NewTable()
	t.Select(0, 0).SetSelectable(true, false).SetTitle(title).SetBorder(true).SetTitleAlign(tview.AlignLeft)
	return t
}

func newTextView(title string) *tview.TextView {
	t := tview.NewTextView()
	t.SetDynamicColors(true).SetTitle(title).SetBorder(true).SetTitleAlign(tview.AlignLeft)
	return t
}

func newInputField() *tview.InputField {
	i := tview.NewInputField()
	i.SetBorder(true).SetTitleAlign(tview.AlignLeft)
	return i
}
