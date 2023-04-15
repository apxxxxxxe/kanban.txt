package tui

import ()

func (t *Tui) setFocusedFunc() {
	t.TodoPane.SetFocusFunc(t.todoPaneInputFocusFunc)
	t.DoingPane.SetFocusFunc(t.doingPaneInputFocusFunc)
	t.DonePane.SetFocusFunc(t.donePaneInputFocusFunc)
}

func (t *Tui) todoPaneInputFocusFunc() {
	t.tableInputFocusFunc(t.TodoPane)
}

func (t *Tui) doingPaneInputFocusFunc() {
	t.tableInputFocusFunc(t.DoingPane)
}

func (t *Tui) donePaneInputFocusFunc() {
	t.tableInputFocusFunc(t.DonePane)
}

func (t *Tui) tableInputFocusFunc(table *TodoTable) {
	table.SetSelectable(true, false)

	row, col := table.GetSelection()
	table.Select(row, col)
}
