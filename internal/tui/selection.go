package tui

import (
	"strconv"

	todo "github.com/1set/todotxt"
)

func (t *Tui) setSelectedFunc() {
	t.TodoPane.SetSelectionChangedFunc(t.todoPaneSelectionChangedFunc)
	t.DoingPane.SetSelectionChangedFunc(t.doingPaneSelectionChangedFunc)
	t.DonePane.SetSelectionChangedFunc(t.donePaneSelectionChangedFunc)
}

func (t *Tui) todoPaneSelectionChangedFunc(row, col int) {
	t.tableSelectionChangedFunc(t.TodoPane, row, col)
}

func (t *Tui) doingPaneSelectionChangedFunc(row, col int) {
	t.tableSelectionChangedFunc(t.DoingPane, row, col)
}

func (t *Tui) donePaneSelectionChangedFunc(row, col int) {
	t.tableSelectionChangedFunc(t.DonePane, row, col)
}

func (t *Tui) tableSelectionChangedFunc(table *TodoTable, row, col int) {
	task, ok := table.GetCell(row, col).Reference.(*todo.Task)
	if ok {
		description := [][]string{
			{"ID", strconv.Itoa(task.ID)},
			{"CreatedDate", task.CreatedDate.Format(todo.DateLayout)},
			{"Priority", task.Priority},
		}
		t.Descript(description)
	} else {
		t.Descript(nil)
	}
}
