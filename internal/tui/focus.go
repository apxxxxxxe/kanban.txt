package tui

import ()

func (t *Tui) setFocusedFunc() {
	t.TodoPane.SetFocusFunc(t.todoPaneInputFocusFunc)
	t.DoingPane.SetFocusFunc(t.doingPaneInputFocusFunc)
	t.DonePane.SetFocusFunc(t.donePaneInputFocusFunc)
  t.DescriptionWidget.SetFocusFunc(t.descriptionWidgetInputFocusFunc)
}

func (t *Tui) todoPaneInputFocusFunc() {
	t.TodoPane.SetTitle(todoPaneTitle)
	t.DoingPane.SetTitle("[l]" + doingPaneTitle)
	t.DonePane.SetTitle(donePaneTitle)
	t.DescriptionWidget.SetTitle("[D]" + descriptionWidgetTitle)
	t.tableInputFocusFunc(t.TodoPane)
}

func (t *Tui) doingPaneInputFocusFunc() {
	t.TodoPane.SetTitle("[h]" + todoPaneTitle)
	t.DoingPane.SetTitle(doingPaneTitle)
	t.DonePane.SetTitle("[l]" + donePaneTitle)
	t.DescriptionWidget.SetTitle("[D]" + descriptionWidgetTitle)
	t.tableInputFocusFunc(t.DoingPane)
}

func (t *Tui) donePaneInputFocusFunc() {
	t.TodoPane.SetTitle(todoPaneTitle)
	t.DoingPane.SetTitle("[h]" + doingPaneTitle)
	t.DonePane.SetTitle(donePaneTitle)
	t.DescriptionWidget.SetTitle("[D]" + descriptionWidgetTitle)
	t.tableInputFocusFunc(t.DonePane)
}

func (t *Tui) tableInputFocusFunc(table *TodoTable) {
	table.SetSelectable(true, false)

	row, col := table.GetSelection()
	table.Select(row, col)
}

func (t *Tui) descriptionWidgetInputFocusFunc() {
  t.DescriptionWidget.SetSelectable(true, false)
}
