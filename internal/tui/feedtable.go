package tui

import (
	todo "github.com/1set/todotxt"
	"github.com/pkg/errors"
	"github.com/rivo/tview"
)

type TodoTable struct {
	*tview.Table
}

var ErrFeedNotExist = errors.Errorf("Feed Not Exist")

func (t *TodoTable) ResetCell(tasklist todo.TaskList) {
	t.Clear()
	for _, task := range tasklist {
		t.setCell(task)
	}
}

func (t *TodoTable) setCell(f todo.Task) *tview.TableCell {
	maxRow := t.GetRowCount()
	targetRow := maxRow
	for i := 0; i < maxRow; i++ {
		cell := t.GetCell(i, 0)
		ref, ok := cell.GetReference().(*FeedCellRef)
		if ok {
			if ref.Feed.ID == f.ID {
				targetRow = i
				break
			}
		}
	}

	cell := tview.NewTableCell(f.Todo).
		SetReference(NewFeedCellRef(f))

	t.SetCell(targetRow, 0, cell)

	// SelectionChangedFuncを発火する
	if maxRow == 0 {
		t.Select(t.GetSelection())
	}

	return cell
}
