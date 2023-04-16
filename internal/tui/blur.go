package tui

import "github.com/rivo/tview"

func (t *Tui) setBlurFunc() {
	t.TodoPane.SetBlurFunc(t.todoPaneBlurFunc)
	t.DoingPane.SetBlurFunc(t.doingPaneBlurFunc)
	t.DonePane.SetBlurFunc(t.donePaneBlurFunc)
	t.DescriptionWidget.SetBlurFunc(t.descriptionWidgetBlurFunc)
}

func (t *Tui) todoPaneBlurFunc() {
	t.tableBlurFunc(t.TodoPane.Table)
}

func (t *Tui) doingPaneBlurFunc() {
	t.tableBlurFunc(t.DoingPane.Table)
}

func (t *Tui) donePaneBlurFunc() {
	t.tableBlurFunc(t.DonePane.Table)
}

func (t *Tui) descriptionWidgetBlurFunc() {
	t.DescriptionWidget.Clear()
	t.tableBlurFunc(t.DescriptionWidget)
}

func (t *Tui) tableBlurFunc(table *tview.Table) {
	table.SetSelectable(false, false)
}
