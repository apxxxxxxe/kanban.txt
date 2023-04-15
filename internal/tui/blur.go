package tui

import ()

func (t *Tui) setBlurFunc() {
	t.TodoPane.SetBlurFunc(t.todoPaneBlurFunc)
	t.DoingPane.SetBlurFunc(t.doingPaneBlurFunc)
	t.DonePane.SetBlurFunc(t.donePaneBlurFunc)
}

func (t *Tui) todoPaneBlurFunc() {
	t.tableBlurFunc(t.TodoPane)
}

func (t *Tui) doingPaneBlurFunc() {
	t.tableBlurFunc(t.DoingPane)
}

func (t *Tui) donePaneBlurFunc() {
	t.tableBlurFunc(t.DonePane)
}

func (t *Tui) tableBlurFunc(table *TodoTable) {
	table.SetSelectable(false, false)
}
