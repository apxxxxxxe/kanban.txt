package tui

import (
	"errors"
	"github.com/1set/todotxt"
	db "github.com/apxxxxxxe/kanban.txt/internal/db"
	tsk "github.com/apxxxxxxe/kanban.txt/internal/task"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
)

const (
	defaultStatus = iota
	todoDelete
	doingDelete
	doneDelete
	taskArchive
)

var ErrReferenceNotFound = errors.New("Reference not found")

func (t *Tui) setKeybind() {
	t.App.SetInputCapture(t.AppInputCaptureFunc)
	t.DaysTable.SetInputCapture(t.daysTableInputCaptureFunc)
	t.ProjectPane.SetInputCapture(t.projectPaneInputCaptureFunc)
	t.TodoPane.SetInputCapture(t.todoPaneInputCaptureFunc)
	t.DoingPane.SetInputCapture(t.doingPaneInputCaptureFunc)
	t.DonePane.SetInputCapture(t.donePaneInputCaptureFunc)
	t.InputWidget.SetInputCapture(t.inputWidgetInputCaptureFunc)
	t.DescriptionWidget.SetInputCapture(t.descriptionWidgetInputCaptureFunc)
}

func (t *Tui) selectTask() (*todotxt.Task, string, error) {
	var cell *tview.TableCell
	if t.TodoPane.HasFocus() {
		cell = t.TodoPane.GetCell(t.TodoPane.GetSelection())
	} else if t.DoingPane.HasFocus() {
		cell = t.DoingPane.GetCell(t.DoingPane.GetSelection())
	} else if t.DonePane.HasFocus() {
		cell = t.DonePane.GetCell(t.DonePane.GetSelection())
	}
	cellText := cell.Text
	task, err := getTaskFromCell(cell)
	return task, cellText, err
}

func getTaskFromCell(cell *tview.TableCell) (*todotxt.Task, error) {
	task, ok := cell.GetReference().(*todotxt.Task)
	if !ok {
		return nil, ErrReferenceNotFound
	}
	return task, nil
}

func (t *Tui) AppInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	if t.InputWidget.HasFocus() {
		return event
	}

	switch event.Rune() {
	case 'q':
		// Quit
		t.App.Stop()
		return nil
	case 'p':
		t.InputWidget.SetTitle("New Project")
		t.Pages.ShowPage(inputField)
		t.pushFocus(t.InputWidget.Box)
		t.InputWidget.Mode = 'p'
		return nil
	case 'n':
		// New task
		t.InputWidget.SetTitle("New Task")
		t.Pages.ShowPage(inputField)
		t.pushFocus(t.InputWidget.Box)
		t.InputWidget.Mode = 'n'
		return nil
	case 'R':
		// Rename Current Project
		t.InputWidget.SetTitle("Rename Project")
		t.Pages.ShowPage(inputField)
		t.pushFocus(t.InputWidget.Box)
		t.InputWidget.Mode = 'R'
		return nil
	case 'a':
		// Archive
		if t.ConfirmationStatus == taskArchive {
			task, _, err := t.selectTask()
			if err != nil {
				t.Notify(err.Error(), true)
				return nil
			}
			t.DB.ArchiveTask(task)
			t.refreshProjects()
			t.Notify("Archived tasks", false)
			t.ConfirmationStatus = defaultStatus
		} else {
			t.ConfirmationStatus = taskArchive
			t.Notify("Press a again to archive all tasks", false)
		}
		return nil
	case 'P':
		// add or increment priority
		task, cellText, err := t.selectTask()
		if err != nil {
			t.Notify(err.Error(), true)
			return nil
		}
		if task != nil {
			if task.Priority == "" {
				task.Priority = priorityA
			} else {
				priorities := []string{
					priorityA,
					priorityB,
					priorityC,
					priorityD,
					priorityE,
				}
				for i, p := range priorities {
					if task.Priority == p {
						if i == len(priorities)-1 {
							task.Priority = ""
						} else {
							task.Priority = priorities[i+1]
						}
						break
					}
				}
			}

			t.refreshProjects()
			if t.TodoPane.HasFocus() {
				t.TodoPane.SelectByText(cellText)
			} else if t.DoingPane.HasFocus() {
				t.DoingPane.SelectByText(cellText)
			} else if t.DonePane.HasFocus() {
				t.DonePane.SelectByText(cellText)
			}
		}
		return nil
	default:
		return event
	}
}

func (t *Tui) daysTableInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'j':
		t.popFocus()
	}
	return event
}

func (t *Tui) moveToDaysTable(table *tview.Table) bool {
	if row, _ := table.GetSelection(); row == 0 {
		t.pushFocus(t.DaysTable.Box)
		return true
	}
	return false
}

func (t *Tui) projectPaneInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'k':
		if t.moveToDaysTable(t.ProjectPane.Table) {
			return nil
		}
	case 'l':
		row, _ := t.TodoPane.GetSelection()
		n := t.TodoPane.GetRowCount()
		if row > n-1 {
			t.TodoPane.Select(n-1, 0)
		}
		t.pushFocus(t.TodoPane.Box)
	}

	return event
}

func (t *Tui) todoPaneInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	f := func() {
		// Move to DoingPane
		project := t.ProjectPane.GetCurrentProject()
		if t.TodoPane.GetRowCount() > 0 && project != nil {
			ref, err := getTaskFromCell(t.TodoPane.GetCell(t.TodoPane.GetSelection()))
			if err != nil {
				panic(err)
			}
			tsk.ToDoing(ref, t.getSelectingDate())
			t.refreshProjects()

			t.TodoPane.AdjustSelection()
		}
	}

	switch event.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		f()
	}

	switch event.Rune() {
	case 'k':
		if t.moveToDaysTable(t.TodoPane.Table) {
			return nil
		}
	case 'd':
		if t.ConfirmationStatus == todoDelete {
			if t.TodoPane.GetRowCount() > 0 {
				task, err := getTaskFromCell(t.TodoPane.GetCell(t.TodoPane.GetSelection()))
				if err != nil {
					panic(err)
				}

				t.DB.LivingTasks.RemoveTask(task)
				t.refreshProjects()

				t.Notify("deleted todo task", false)
			} else {
				t.Notify("No todo task here", true)
			}
			t.ConfirmationStatus = defaultStatus
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
		t.pushFocus(t.ProjectPane.Box)
	case 'l':
		row, _ := t.DoingPane.GetSelection()
		n := t.DoingPane.GetRowCount()
		if row > n-1 {
			t.DoingPane.Select(n-1, 0)
		}
		t.pushFocus(t.DoingPane.Box)
	case 'J':
		t.EditingCell = t.TodoPane.GetCell(t.TodoPane.GetSelection())
		t.pushFocus(t.DescriptionWidget.Box)
	case ' ':
		f()
	}

	return event
}

func (t *Tui) doingPaneInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		// Move to TodoPane
		ref, err := getTaskFromCell(t.DoingPane.GetCell(t.DoingPane.GetSelection()))
		if err != nil {
			panic("doingPaneInputCaptureFunc: ref is not todotxt.Task")
		}
		tsk.ToTodo(ref)
		t.refreshProjects()

		t.DoingPane.AdjustSelection()
	}

	switch event.Rune() {
	case 'k':
		if t.moveToDaysTable(t.DoingPane.Table) {
			return nil
		}
	case 'd':
		if t.ConfirmationStatus == doingDelete {
			if t.DoingPane.GetRowCount() > 0 {
				task, err := getTaskFromCell(t.DoingPane.GetCell(t.DoingPane.GetSelection()))
				if err != nil {
					panic(err)
				}

				t.DB.LivingTasks.RemoveTask(task)
				t.refreshProjects()

				t.Notify("Deleted doing task", false)
			} else {
				t.Notify("No doing task here", true)
			}
			t.ConfirmationStatus = defaultStatus
		} else {
			t.ConfirmationStatus = doingDelete
			t.Notify("Press d again to delete doing task", false)
		}
		return event
	case ' ':
		// Move to DonePane
		if t.DoingPane.GetRowCount() > 0 {
			ref, err := getTaskFromCell(t.DoingPane.GetCell(t.DoingPane.GetSelection()))
			if err != nil {
				panic(err)
			}
			tsk.ToDone(ref, t.getSelectingDate())
			t.refreshProjects()

			t.DoingPane.AdjustSelection()
		}
	case 'h':
		t.pushFocus(t.TodoPane.Box)
	case 'l':
		t.pushFocus(t.DonePane.Box)
	case 'J':
		t.EditingCell = t.DoingPane.GetCell(t.DoingPane.GetSelection())
		t.pushFocus(t.DescriptionWidget.Box)
	}

	return event
}

func (t *Tui) donePaneInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	f := func() {
		// Move to DoingPane
		if t.DonePane.GetRowCount() > 0 {
			ref, err := getTaskFromCell(t.DonePane.GetCell(t.DonePane.GetSelection()))
			if err != nil {
				panic(err)
			}
			tsk.ToDoing(ref, t.getSelectingDate())
			t.refreshProjects()

			t.DonePane.AdjustSelection()
		}
	}

	switch event.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		f()
	}

	switch event.Rune() {
	case 'k':
		if t.moveToDaysTable(t.DonePane.Table) {
			return nil
		}
	case 'd':
		if t.ConfirmationStatus == doneDelete {
			if t.DonePane.GetRowCount() > 0 {
				task, err := getTaskFromCell(t.DonePane.GetCell(t.DonePane.GetSelection()))
				if err != nil {
					panic(err)
				}

				t.DB.LivingTasks.RemoveTask(task)
				t.refreshProjects()

				t.Notify("Deleted done task", false)
			} else {
				t.Notify("No done task here", true)
			}
			t.ConfirmationStatus = defaultStatus
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
		t.pushFocus(t.DoingPane.Box)
	case 'J':
		t.EditingCell = t.DonePane.GetCell(t.DonePane.GetSelection())
		t.pushFocus(t.DescriptionWidget.Box)
	}

	return event
}

func (t *Tui) inputWidgetInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	project := t.ProjectPane.GetCurrentProject()

	hideInputField := func() {
		focus := t.App.GetFocus()
		t.Pages.HidePage(inputField)

		// it prevents the focus from the effect of HidePage
		t.App.SetFocus(focus)

		t.InputWidget.SetText("")
		t.InputWidget.Mode = ' '
	}

	switch event.Key() {
	case tcell.KeyEscape:
		t.Pages.HidePage(inputField)
		t.popFocus()
		t.InputWidget.SetText("")
		switch t.InputWidget.Mode {
		case 'f':
			t.popFocus()
		}
		t.InputWidget.Mode = ' '
		return nil
	case tcell.KeyEnter:
		input := t.InputWidget.GetText()
		switch t.InputWidget.Mode {
		case 'n':
			// New Task
			taskFields := []string{}
			for _, field := range strings.Split(input, " ") {
				if field == "" {
					continue
				}
				taskFields = append(taskFields, tsk.ReplaceInvalidTag(field)+" ")
			}
			task, err := todotxt.ParseTask(strings.Join(taskFields, " "))
			if err != nil {
				t.Notify(err.Error(), true)
				return nil
			}

			// add CreatedDate
			task.CreatedDate = t.getSelectingDate()

			// add current project
			if project == nil || project.ProjectName == db.AllTasks {
				task.Projects = []string{db.NoProject}
			} else {
				task.Projects = []string{project.ProjectName}
			}

			// remove context "doing"
			for i, context := range task.Contexts {
				if context == "doing" {
					task.Contexts = append(task.Contexts[:i], task.Contexts[i+1:]...)
					break
				}
			}

			t.DB.LivingTasks.AddTask(task)
			t.refreshProjects()

		case 'p':
			// New Project
			day, _ := t.getCurrentDay()
			t.DB.RefreshProjects(day)
			t.DB.Projects = append(t.DB.Projects, &db.Project{ProjectName: input})
			t.ProjectPane.ResetCell(t.DB.Projects)

		case 'R':
			// Rename Project
			taskList := t.DB.LivingTasks.Filter(todotxt.FilterByProject(project.ProjectName))
			for _, task := range *taskList {
				task.Projects = []string{input}
			}
			t.refreshProjects()

		case 'f':
			// Edit Field
			field := t.InputWidget.GetTitle()
			cellText := t.EditingCell.Text
			task, err := getTaskFromCell(t.EditingCell)
			if err != nil {
				panic(err)
			}
			setTaskField(task, field, input)

			t.popFocus() // pop focus from inputWidget
			t.popFocus() // pop focus from descriptionWidget
			// now focus is on the pane

			var selectCell func(string)
			if t.TodoPane.HasFocus() {
				selectCell = t.TodoPane.SelectByText
			} else if t.DoingPane.HasFocus() {
				selectCell = t.DoingPane.SelectByText
			} else if t.DonePane.HasFocus() {
				selectCell = t.DonePane.SelectByText
			} else {
				panic("inputWidgetInputCaptureFunc: no pane has focus")
			}

			t.refreshProjects()
			hideInputField()
			selectCell(cellText)
			return nil
		}

		t.popFocus()
		hideInputField()
		return nil
	}

	return event
}

func (t *Tui) descriptionWidgetInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	f := func() {
		row, _ := t.DescriptionWidget.GetSelection()
		field := t.DescriptionWidget.GetCell(row, 0).GetReference().(string)
		t.InputWidget.SetTitle(field)
		task, ok := t.EditingCell.GetReference().(*todotxt.Task)
		if !ok {
			panic("descriptionWidgetInputCaptureFunc: ref is not *todotxt.Task")
		}
		t.InputWidget.SetText(getTaskField(task, field))
		t.InputWidget.Mode = 'f'
		t.Pages.ShowPage(inputField)
		t.pushFocus(t.InputWidget.Box)
	}

	switch event.Key() {
	case tcell.KeyEnter:
		f()
	}
	switch event.Rune() {
	case 'K':
		t.popFocus()
	case ' ':
		f()
	}

	return event
}
