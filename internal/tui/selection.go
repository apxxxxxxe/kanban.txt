package tui

import ()

func (t *Tui) setSelectedFunc() {
	t.TodoPane.SetSelectionChangedFunc(t.todoPaneSelectionChangedFunc)
	t.DoingPane.SetSelectionChangedFunc(t.doingPaneSelectionChangedFunc)
	t.DonePane.SetSelectionChangedFunc(t.donePaneSelectionChangedFunc)
}

func (t *Tui) todoPaneSelectionChangedFunc(row, col int) {
}

func (t *Tui) doingPaneSelectionChangedFunc(row, col int) {
}

func (t *Tui) donePaneSelectionChangedFunc(row, col int) {
}
