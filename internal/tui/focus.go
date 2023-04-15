package tui

import ()

func (t *Tui) setFocusedFunc() {
	t.TodoPane.SetFocusFunc(t.todoPaneInputFocusFunc)
	t.DoingPane.SetFocusFunc(t.doingPaneInputFocusFunc)
	t.DonePane.SetFocusFunc(t.donePaneInputFocusFunc)
}

func (t *Tui) todoPaneInputFocusFunc() {
}

func (t *Tui) doingPaneInputFocusFunc() {
}

func (t *Tui) donePaneInputFocusFunc() {
}
