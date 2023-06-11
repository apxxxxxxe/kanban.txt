package tui

import (
	"fmt"
	"time"

	todo "github.com/1set/todotxt"
	db "github.com/apxxxxxxe/kanban.txt/internal/db"
)

func (t *Tui) setSelectedFunc() {
	t.ProjectPane.SetSelectionChangedFunc(t.projectPaneSelectionChangedFunc)
	t.TodoPane.SetSelectionChangedFunc(t.todoPaneSelectionChangedFunc)
	t.DoingPane.SetSelectionChangedFunc(t.doingPaneSelectionChangedFunc)
	t.DonePane.SetSelectionChangedFunc(t.donePaneSelectionChangedFunc)
}

func (t *Tui) projectPaneSelectionChangedFunc(row, col int) {
	if t.ProjectPane.GetRowCount() == 0 {
		return
	}

	project, ok := t.ProjectPane.GetCell(row, col).Reference.(*db.Project)
	if !ok {
		panic("invalid reference")
	}

	t.TodoPane.ResetCell(project.TodoTasks)
	t.DoingPane.ResetCell(project.DoingTasks)
	t.DonePane.ResetCell(project.DoneTasks)

	description := [][]string{
		{"wholetasklen", fmt.Sprintf("%d", len(t.DB.WholeTasks))},
		{"name", project.ProjectName},
		{"len", fmt.Sprintf("%d", len(project.TodoTasks)+len(project.DoingTasks)+len(project.DoneTasks))},
		{"todo", project.TodoTasks.String()},
		{"doing", project.DoingTasks.String()},
		{"done", project.DoneTasks.String()},
	}
	t.Descript(description)
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
		contexts := ""
		if task.Contexts != nil {
			for _, c := range task.Contexts {
				contexts += c + " "
			}
		}
		projects := ""
		if task.Projects != nil {
			for _, p := range task.Projects {
				projects += p + " "
			}
		}
		description := [][]string{
			// {"ID", strconv.Itoa(task.ID)},
			{"Projects", projects},
			{"Priority", task.Priority},
			{"Title", task.Todo},
			{"Contexts", contexts},
			{"DueDate", timeToStr(task.DueDate)},
			{"CompletedDate", timeToStr(task.CompletedDate)},
			{"CreatedDate", timeToStr(task.CreatedDate)},
		}
		t.Descript(description)
	} else {
		t.Descript(nil)
	}
}

func timeToStr(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(todo.DateLayout)
}
