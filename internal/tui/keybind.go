package tui

import (
	todo "github.com/1set/todotxt"
	db "github.com/apxxxxxxe/kanban.txt/internal/db"
	"github.com/gdamore/tcell/v2"
)

const (
	defaultStatus = iota
	todoDelete
	doingDelete
	doneDelete
)

func (t *Tui) setKeybind() {
	t.App.SetInputCapture(t.AppInputCaptureFunc)
	t.ProjectPane.SetInputCapture(t.projectPaneInputCaptureFunc)
	t.TodoPane.SetInputCapture(t.todoPaneInputCaptureFunc)
	t.DoingPane.SetInputCapture(t.doingPaneInputCaptureFunc)
	t.DonePane.SetInputCapture(t.donePaneInputCaptureFunc)
	t.InputWidget.SetInputCapture(t.inputWidgetInputCaptureFunc)
}

func (t *Tui) AppInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'q':
		// Quit
		t.App.Stop()
		return nil
	case 'n':
		// New task
		switch t.App.GetFocus() {
		case t.InputWidget:
			return event
		case t.ProjectPane:
			t.InputWidget.SetTitle("New Project")
			t.Pages.ShowPage(inputField)
			t.App.SetFocus(t.InputWidget)
			t.InputWidget.Mode = 'p'
			return nil
		default:
			t.InputWidget.SetTitle("New Task")
			t.Pages.ShowPage(inputField)
			t.App.SetFocus(t.InputWidget)
			t.InputWidget.Mode = 'n'
			return nil
		}
	}

	return event
}

func (t *Tui) projectPaneInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'l':
		row, _ := t.TodoPane.GetSelection()
		n := t.TodoPane.GetRowCount()
		if row > n-1 {
			t.TodoPane.Select(n-1, 0)
		}
		t.App.SetFocus(t.TodoPane)
	}

	return event
}

func (t *Tui) deleteTask(pane *TodoTable) {
	cell := pane.GetCell(pane.GetSelection())
	ref, _ := cell.GetReference().(*todo.Task)

	var taskList *todo.TaskList
	switch pane.GetTitle() {
	case todoPaneTitle:
		taskList = &t.ProjectPane.GetCurrentProject().TodoTasks
	case doingPaneTitle:
		taskList = &t.ProjectPane.GetCurrentProject().DoingTasks
	case donePaneTitle:
		taskList = &t.ProjectPane.GetCurrentProject().DoneTasks
	}

	if err := taskList.RemoveTaskByID(ref.ID); err != nil {
		panic(err)
	}
	if err := t.DB.SaveData(); err != nil {
		panic(err)
	}
	pane.ResetCell(taskList)
}

func (t *Tui) todoPaneInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	project := t.ProjectPane.GetCurrentProject()

	f := func() {
		// Move to DoingPane
		if t.TodoPane.GetRowCount() > 0 {
			row, _ := t.TodoPane.GetSelection()
			cell := t.TodoPane.GetCell(row, 0)
			ref, _ := cell.GetReference().(*todo.Task)
			ref.Contexts = []string{"doing"}
			project.TodoTasks.RemoveTaskByID(ref.ID)
			project.DoingTasks.AddTask(ref)

			if err := t.DB.SaveData(); err != nil {
				panic(err)
			}

			t.TodoPane.ResetCell(&project.TodoTasks)
			t.DoingPane.ResetCell(&project.DoingTasks)
			t.TodoPane.AdjustSelection()
		}
	}

	switch event.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		f()
	}

	switch event.Rune() {
	case 'd':
		if t.ConfirmationStatus == todoDelete {
			t.deleteTask(t.TodoPane)
		} else {
			t.ConfirmationStatus = todoDelete
			t.Notify("Press d again to delete todo task", false)
		}
		return event
	case 'h':
		row, _ := t.ProjectPane.GetSelection()
		n := t.ProjectPane.GetRowCount()
		if row > n-1 {
			t.ProjectPane.Select(n-1, 0)
		}
		t.App.SetFocus(t.ProjectPane)
	case 'l':
		row, _ := t.DoingPane.GetSelection()
		n := t.DoingPane.GetRowCount()
		if row > n-1 {
			t.DoingPane.Select(n-1, 0)
		}
		t.App.SetFocus(t.DoingPane)
	case ' ':
		f()
	}

	return event
}

func (t *Tui) doingPaneInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	project := t.ProjectPane.GetCurrentProject()

	switch event.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		// Move to TodoPane
		row, _ := t.DoingPane.GetSelection()
		cell := t.DoingPane.GetCell(row, 0)
		ref, _ := cell.GetReference().(*todo.Task)
		project.DoingTasks.RemoveTaskByID(ref.ID)
		ref.Contexts = []string{}
		project.TodoTasks.AddTask(ref)

		if err := t.DB.SaveData(); err != nil {
			panic(err)
		}

		t.DoingPane.ResetCell(&project.DoingTasks)
		t.TodoPane.ResetCell(&project.TodoTasks)
		t.DoingPane.AdjustSelection()
	}

	switch event.Rune() {
	case 'd':
		if t.ConfirmationStatus == doingDelete {
			t.deleteTask(t.DoingPane)
		} else {
			t.ConfirmationStatus = doingDelete
			t.Notify("Press d again to delete doing task", false)
		}
		return event
	case ' ':
		// Move to DonePane
		if t.DoingPane.GetRowCount() > 0 {
			row, _ := t.DoingPane.GetSelection()
			cell := t.DoingPane.GetCell(row, 0)
			ref, _ := cell.GetReference().(*todo.Task)
			ref.Contexts = []string{}
			ref.Complete()
			project.DoingTasks.RemoveTaskByID(ref.ID)
			project.DoneTasks.AddTask(ref)

			if err := t.DB.SaveData(); err != nil {
				panic(err)
			}

			t.DoingPane.ResetCell(&project.DoingTasks)
			t.DonePane.ResetCell(&project.DoneTasks)
			t.DoingPane.AdjustSelection()
		}
	case 'h':
		t.App.SetFocus(t.TodoPane)
	case 'l':
		t.App.SetFocus(t.DonePane)
	}

	return event
}

func (t *Tui) donePaneInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	project := t.ProjectPane.GetCurrentProject()

	f := func() {
		// Move to DoingPane
		if t.DonePane.GetRowCount() > 0 {
			row, _ := t.DonePane.GetSelection()
			cell := t.DonePane.GetCell(row, 0)
			ref, _ := cell.GetReference().(*todo.Task)
			ref.Contexts = []string{"doing"}
			ref.Reopen()
			project.DoneTasks.RemoveTaskByID(ref.ID)
			project.DoingTasks.AddTask(ref)

			if err := t.DB.SaveData(); err != nil {
				panic(err)
			}

			t.DonePane.ResetCell(&project.DoneTasks)
			t.DoingPane.ResetCell(&project.DoingTasks)
			t.DonePane.AdjustSelection()
		}
	}

	switch event.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		f()
	}

	switch event.Rune() {
	case 'd':
		if t.ConfirmationStatus == doneDelete {
			t.deleteTask(t.DoingPane)
		} else {
			t.ConfirmationStatus = doneDelete
			t.Notify("Press d again to delete done task", false)
		}
		return event
	case ' ':
		f()
		// or Move to Archive
	case 'h':
		row, _ := t.DoingPane.GetSelection()
		n := t.DoingPane.GetRowCount()
		if row > n-1 {
			t.DoingPane.Select(n-1, 0)
		}
		t.App.SetFocus(t.DoingPane)
	}

	return event
}

func (t *Tui) inputWidgetInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	project := t.ProjectPane.GetCurrentProject()

	switch event.Key() {
	case tcell.KeyEnter:
		input := t.InputWidget.GetText()
		switch t.InputWidget.Mode {
		case 'n':
			// New Task
			task, err := todo.ParseTask(input)
			if err != nil {
				t.Notify(err.Error(), true)
				return nil
			}

			// add current project
			task.Projects = []string{project.ProjectName}

			// remove context "doing"
			for i, context := range task.Contexts {
				if context == "doing" {
					task.Contexts = append(task.Contexts[:i], task.Contexts[i+1:]...)
					break
				}
			}

			task.Reopen()

			project.TodoTasks.AddTask(task)
			t.TodoPane.ResetCell(&project.TodoTasks)
			if err := t.DB.SaveData(); err != nil {
				panic(err)
			}
			t.Pages.HidePage(inputField)
			t.App.SetFocus(t.TodoPane)

		case 'p':
			// New Project
			t.DB.Projects = append(t.DB.Projects, &db.Project{ProjectName: input})
			t.ProjectPane.ResetCell(t.DB.Projects)
			t.Pages.HidePage(inputField)
			t.App.SetFocus(t.ProjectPane)
		}

		t.InputWidget.SetText("")
		t.InputWidget.Mode = ' '
		return nil
	}

	return event
}
