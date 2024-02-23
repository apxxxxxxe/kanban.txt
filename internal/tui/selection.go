package tui

import (
	"fmt"

	"github.com/1set/todotxt"
	db "github.com/apxxxxxxe/kanban.txt/internal/db"
)

func (t *Tui) setSelectedFunc() {
	t.DaysTable.SetSelectionChangedFunc(t.daysTableSelectionChangedFunc)
	t.ProjectPane.SetSelectionChangedFunc(t.projectPaneSelectionChangedFunc)
	t.TodoPane.SetSelectionChangedFunc(t.todoPaneSelectionChangedFunc)
	t.DoingPane.SetSelectionChangedFunc(t.doingPaneSelectionChangedFunc)
	t.DonePane.SetSelectionChangedFunc(t.donePaneSelectionChangedFunc)
}

func (t *Tui) reDrawProjects() (*db.Project, int) {
	project, ok := t.ProjectPane.GetCell(t.ProjectPane.GetSelection()).Reference.(*db.Project)
	if !ok {
		panic("invalid reference")
	}

	_, index := t.getCurrentDay()
	tasks := project.TasksByDate[index]

	t.TodoPane.ResetCell(tasks.TodoTasks)
	t.DoingPane.ResetCell(tasks.DoingTasks)
	t.DonePane.ResetCell(tasks.DoneTasks)

	return project, index
}

func (t *Tui) daysTableSelectionChangedFunc(row, col int) {
	t.reDrawProjects()
}

func (t *Tui) projectPaneSelectionChangedFunc(row, col int) {
	if t.ProjectPane.GetRowCount() == 0 {
		return
	}

	project, index := t.reDrawProjects()
	tasks := project.TasksByDate[index]

	description := [][]string{
		{"wholetasklen", fmt.Sprintf("%d", len(t.DB.LivingTasks))},
		{"name", project.ProjectName},
		{"len", fmt.Sprintf("%d", len(tasks.TodoTasks)+len(tasks.DoingTasks)+len(tasks.DoneTasks))},
		{"todo", fmt.Sprintf("%d", len(tasks.TodoTasks))},
		{"doing", fmt.Sprintf("%d", len(tasks.DoingTasks))},
		{"done", fmt.Sprintf("%d", len(tasks.DoneTasks))},
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
			todoDueDate,
			todoCompletedDate,
			todoCreatedDate,
			todoRecurrence,
			todoNote,
		}
		description := [][]string{}
		for _, field := range fields {
			description = append(description, []string{field, getTaskField(task, field)})
		}
		t.Descript(description)
	} else {
		t.Descript(nil)
	}
}
