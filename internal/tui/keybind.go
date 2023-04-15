package tui

import (
	todo "github.com/1set/todotxt"
	"github.com/gdamore/tcell/v2"
)

func (t *Tui) setKeybind() {
	t.TodoPane.SetInputCapture(t.todoPaneInputCaptureFunc)
	t.DoingPane.SetInputCapture(t.doingPaneInputCaptureFunc)
	t.DonePane.SetInputCapture(t.donePaneInputCaptureFunc)
}

func (t *Tui) todoPaneInputCaptureFunc(event *tcell.EventKey) *tcell.EventKey {
	f := func() {
		// Move to DoingPane
		if t.TodoPane.GetRowCount() > 0 {
			row, _ := t.TodoPane.GetSelection()
			cell := t.TodoPane.GetCell(row, 0)
			ref, _ := cell.GetReference().(*todo.Task)
			t.DB.TodoTasks.RemoveTaskByID(ref.ID)
			t.DB.DoingTasks.AddTask(ref)
			t.TodoPane.ResetCell(t.DB.TodoTasks)
			t.DoingPane.ResetCell(t.DB.DoingTasks)
			t.TodoPane.AdjustSelection()
		}
	}

	switch event.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		f()
	}

	switch event.Rune() {
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
	switch event.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		// Move to TodoPane
		row, _ := t.DoingPane.GetSelection()
		cell := t.DoingPane.GetCell(row, 0)
		ref, _ := cell.GetReference().(*todo.Task)
		t.DB.DoingTasks.RemoveTaskByID(ref.ID)
		t.DB.TodoTasks.AddTask(ref)
		t.DoingPane.ResetCell(t.DB.DoingTasks)
		t.TodoPane.ResetCell(t.DB.TodoTasks)
		t.DoingPane.AdjustSelection()
	}

	switch event.Rune() {
	case ' ':
		// Move to DonePane
		if t.DoingPane.GetRowCount() > 0 {
			row, _ := t.DoingPane.GetSelection()
			cell := t.DoingPane.GetCell(row, 0)
			ref, _ := cell.GetReference().(*todo.Task)
			t.DB.DoingTasks.RemoveTaskByID(ref.ID)
			t.DB.DoneTasks.AddTask(ref)
			t.DoingPane.ResetCell(t.DB.DoingTasks)
			t.DonePane.ResetCell(t.DB.DoneTasks)
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
	f := func() {
		// Move to DoingPane
		if t.DonePane.GetRowCount() > 0 {
			row, _ := t.DonePane.GetSelection()
			cell := t.DonePane.GetCell(row, 0)
			ref, _ := cell.GetReference().(*todo.Task)
			t.DB.DoneTasks.RemoveTaskByID(ref.ID)
			t.DB.DoingTasks.AddTask(ref)
			t.DonePane.ResetCell(t.DB.DoneTasks)
			t.DoingPane.ResetCell(t.DB.DoingTasks)
			t.DonePane.AdjustSelection()
		}
	}

	switch event.Key() {
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		f()
	}

	switch event.Rune() {
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
