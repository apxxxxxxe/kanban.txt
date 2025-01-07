package tui

import (
	"github.com/1set/todotxt"
	"github.com/rivo/tview"
)

func (t *Tui) setSelectedFunc() {
	t.DaysTable.SetSelectionChangedFunc(t.daysTableSelectionChangedFunc)
	t.ProjectPane.SetSelectionChangedFunc(t.projectPaneSelectionChangedFunc)
	t.TodoPane.SetSelectionChangedFunc(t.todoPaneSelectionChangedFunc)
	t.DoingPane.SetSelectionChangedFunc(t.doingPaneSelectionChangedFunc)
	t.DonePane.SetSelectionChangedFunc(t.donePaneSelectionChangedFunc)
}

func (t *Tui) reDrawProjects() {
	day, _ := t.getCurrentDay()
	t.DB.RefreshProjects(day)
	projects := t.DB.Projects

	// TODO: 見た目との分離; 現在はProjectsByDateの各要素間でProjectとその並びが同一であることを前提にしている
	projectIndex, _ := t.ProjectPane.GetSelection()
	if len(projects) > 0 {
		project := projects[projectIndex]

		t.TodoPane.ResetCell(project.TodoTasks)
		t.DoingPane.ResetCell(project.DoingTasks)
		t.DonePane.ResetCell(project.DoneTasks)
	}
}

func (t *Tui) daysTableSelectionChangedFunc(row, col int) {
	t.reDrawProjects()
}

func (t *Tui) projectPaneSelectionChangedFunc(row, col int) {
	if t.ProjectPane.GetRowCount() != 0 {
		t.reDrawProjects()
	}
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
	task, ok := table.GetCell(row, col).Reference.(*todotxt.Task)
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
		fields := []string{
			todoProjects,
			todoPriority,
			todoTitle,
			todoContexts,
			todoCreatedDate,
			todoMakedDoing,
			todoDueDate,
			todoCompletedDate,
			todoRecurrence,
			todoNote,
		}
		description := [][]string{}
		for _, field := range fields {
			description = append(description, []string{field, tview.Escape(getTaskField(task, field))})
		}
		t.Descript(description)
	} else {
		t.Descript(nil)
	}
}
