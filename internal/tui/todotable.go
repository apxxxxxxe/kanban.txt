package tui

import (
	todo "github.com/1set/todotxt"
	"github.com/gdamore/tcell/v2"
	"github.com/pkg/errors"
	"github.com/rivo/tview"

	db "github.com/apxxxxxxe/kanban.txt/internal/db"
)

const (
	priorityA = "A"
	priorityB = "B"
	priorityC = "C"
	priorityD = "D"
	priorityE = "E"
	colorA    = tcell.ColorRed
	colorB    = tcell.ColorOrange
	colorC    = tcell.ColorYellow
	colorD    = tcell.ColorGreen
	colorE    = tcell.ColorBlue
)

type TodoTable struct {
	*tview.Table
}

var ErrFeedNotExist = errors.Errorf("Feed Not Exist")

func (t *TodoTable) AdjustSelection() {
	n := t.GetRowCount()
	if n == 0 {
		return
	}

	if row, _ := t.GetSelection(); row > n-1 {
		t.Select(n-1, 0)
	}
}

func (t *TodoTable) ResetCell(tasklist db.TaskReferences) {
	t.Clear()
	for _, task := range tasklist {
		t.setCell(task)
	}
}

func (t *TodoTable) setCell(f *todo.Task) *tview.TableCell {
	maxRow := t.GetRowCount()
	targetRow := maxRow
	for i := 0; i < maxRow; i++ {
		cell := t.GetCell(i, 0)
		ref, ok := cell.GetReference().(*todo.Task)
		if ok {
			if ref.String() == f.String() {
				targetRow = i
				break
			}
		}
	}

	cell := tview.NewTableCell(f.Todo).SetReference(f)

	if f.HasPriority() {
		switch f.Priority {
		case priorityA:
			cell.SetTextColor(colorA)
		case priorityB:
			cell.SetTextColor(colorB)
		case priorityC:
			cell.SetTextColor(colorC)
		case priorityD:
			cell.SetTextColor(colorD)
		case priorityE:
			cell.SetTextColor(colorE)
		}
	}

	if f.Completed {
		cell.SetTextColor(tcell.ColorGray)
	}

	t.SetCell(targetRow, 0, cell)

	// SelectionChangedFuncを発火する
	if maxRow == 0 {
		t.Select(t.GetSelection())
	}

	return cell
}
