package tui

import (
	db "github.com/apxxxxxxe/kanban.txt/internal/db"
	"github.com/rivo/tview"
)

type ProjectTable struct {
	*tview.Table
}

func (t *ProjectTable) GetCurrentProject() *db.Project {
	p, ok := t.GetCell(t.GetSelection()).GetReference().(*db.Project)
	if !ok {
		panic("invalid reference")
	}
	return p
}

func (t *ProjectTable) ResetCell(projects []*db.Project) {
	t.Clear()
	for _, project := range projects {
		t.setCell(project)
	}
}

func (t *ProjectTable) setCell(p *db.Project) *tview.TableCell {
	maxRow := t.GetRowCount()
	targetRow := maxRow
	for i := 0; i < maxRow; i++ {
		cell := t.GetCell(i, 0)
		ref, ok := cell.GetReference().(*db.Project)
		if ok {
			if ref.ProjectName == p.ProjectName {
				targetRow = i
				break
			}
		}
	}

	cell := tview.NewTableCell(p.ProjectName).SetReference(p)

	t.SetCell(targetRow, 0, cell)

	// SelectionChangedFuncを発火する
	if maxRow == 0 {
		t.Select(t.GetSelection())
	}

	return cell
}
