package tui

import (
	"fmt"
	db "github.com/apxxxxxxe/kanban.txt/internal/db"
	"github.com/apxxxxxxe/kanban.txt/pkg/util"
	"github.com/pkg/errors"
	"github.com/rivo/tview"
	"time"
)

type Tui struct {
	Config             *db.Config
	DB                 *db.Database
	App                *tview.Application
	Pages              *tview.Pages
	DaysTable          *tview.Table
	ProjectPane        *ProjectTable
	TodoPane           *TodoTable
	DoingPane          *TodoTable
	DonePane           *TodoTable
	DescriptionWidget  *tview.Table
	InfoWidget         *tview.TextView
	HelpWidget         *tview.TextView
	InputWidget        *InputBox
	ColorWidget        *tview.Table
	FocusStack         []*tview.Box
	EditingCell        *tview.TableCell
	ConfirmationStatus int
	CurrentLeftTable   int
	IsLoading          bool
}

const (
	descriptionField       = "descPopup"
	inputField             = "InputPopup"
	colorTable             = "ColorTablePopup"
	mainPage               = "MainPage"
	keymapPage             = "KeymapPage"
	projectPaneTitle       = "Project"
	todoPaneTitle          = "Todo"
	doingPaneTitle         = "Doing"
	donePaneTitle          = "Done"
	descriptionWidgetTitle = "Description"
	helpWidgetTitle        = "Help"
	infoWidgetTitle        = "Info"
	colorWidgetTitle       = "Color"
)

const (
	enumTodoPane = iota
	enumDoingPane
)

var ErrImportFileNotFound = errors.Errorf("todo.txt not found")

func NewTui() *Tui {
	tview.Styles.ContrastBackgroundColor = tview.Styles.PrimitiveBackgroundColor

	tui := &Tui{
		Config:             db.LoadOrNewConfig(),
		DB:                 &db.Database{},
		App:                tview.NewApplication(),
		Pages:              tview.NewPages(),
		DaysTable:          tview.NewTable().SetBorders(false).SetSelectable(false, true),
		ProjectPane:        &ProjectTable{newTable(projectPaneTitle)},
		TodoPane:           &TodoTable{newTable(todoPaneTitle)},
		DoingPane:          &TodoTable{newTable(doingPaneTitle)},
		DonePane:           &TodoTable{newTable(donePaneTitle)},
		DescriptionWidget:  newTable(descriptionWidgetTitle),
		InfoWidget:         newTextView(infoWidgetTitle),
		HelpWidget:         newTextView(helpWidgetTitle).SetTextAlign(1).SetDynamicColors(true),
		InputWidget:        &InputBox{InputField: newInputField(), Mode: 0},
		FocusStack:         []*tview.Box{},
		EditingCell:        nil,
		ConfirmationStatus: defaultStatus,
		CurrentLeftTable:   enumTodoPane,
		IsLoading:          false,
	}

	const NonZero = 1
	daysLabels := []string{}
	now := time.Now()
	for i := -db.DayCount / 2; i <= db.DayCount/2; i++ {
		daysLabels = append(daysLabels, now.AddDate(0, 0, i).Format("2006-01-02"))
	}
	for i, label := range daysLabels {
		tui.DaysTable.SetCell(0, i, tview.NewTableCell(label).SetAlign(tview.AlignCenter).SetExpansion(NonZero))
	}
	tui.DaysTable.Select(0, db.DayCount/2)

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tui.DaysTable, 1, 0, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(tui.ProjectPane, 0, 1, false).
			AddItem(
				tview.NewFlex().SetDirection(tview.FlexRow).
					AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
						AddItem(tui.TodoPane, 0, 1, false).
						AddItem(tui.DoingPane, 0, 1, false).
						AddItem(tui.DonePane, 0, 1, false),
						0, 3, false).
					AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
						AddItem(tui.DescriptionWidget, 0, 2, false).
						AddItem(tui.InfoWidget, 0, 1, false),
						0, 1, false), 0, 3, false), 0, 1, false)

	inputFlex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(tui.InputWidget, 3, 1, false).
			AddItem(nil, 0, 1, false), 40, 1, false).
		AddItem(nil, 0, 1, false)

	tui.Pages.
		AddPage(mainPage, mainFlex, true, true).
		AddPage(inputField, inputFlex, true, false)

	tui.App.SetRoot(tui.Pages, true)

	tui.setKeybind()
	tui.setSelectedFunc()
	tui.setFocusedFunc()
	tui.setBlurFunc()

	return tui
}

func (t *Tui) pushFocus(b *tview.Box) {
	t.FocusStack = append(t.FocusStack, b)
	t.App.SetFocus(b)
}

func (t *Tui) popFocus() {
	if len(t.FocusStack) == 0 {
		return
	}
	t.FocusStack = t.FocusStack[:len(t.FocusStack)-1]
	if len(t.FocusStack) == 0 {
		return
	}
	t.App.SetFocus(t.FocusStack[len(t.FocusStack)-1])
}

func (t *Tui) Descript(desc [][]string) {
	t.DescriptionWidget.Clear()
	if desc == nil {
		return
	}
	for i, line := range desc {
		titleCell := tview.NewTableCell("[#a0a0a0::b]" + line[0])
		titleCell.SetReference(line[0])
		t.DescriptionWidget.SetCell(i, 0, titleCell)
		t.DescriptionWidget.SetCellSimple(i, 1, line[1])
	}
}

func (t *Tui) getSelectingDate() time.Time {
	_, col := t.DaysTable.GetSelection()
	date := time.Now().AddDate(0, 0, col-db.DayCount/2)
	return util.RemoveClockTime(date)
}

func (t *Tui) getCurrentDay() (int, int) {
	_, col := t.DaysTable.GetSelection()
	return col - db.DayCount/2, col
}

func (t *Tui) refreshProjects() {
  day, _ := t.getCurrentDay()
	if err := t.DB.RefreshProjects(day); err != nil {
		t.Notify(err.Error(), true)
	}
	row, col := t.ProjectPane.GetSelection()
	t.ProjectPane.ResetCell(t.DB.Projects)
	if row >= t.ProjectPane.GetRowCount() {
		row = t.ProjectPane.GetRowCount() - 1
	}
	if col >= t.ProjectPane.GetColumnCount() {
		col = t.ProjectPane.GetColumnCount() - 1
	}
	t.ProjectPane.Select(row, col)
	t.App.SetFocus(t.App.GetFocus())
}

func (t *Tui) Notify(m string, red bool) {
	if red {
		m = "[#ff0000::b]" + m
	}
	t.InfoWidget.SetText(m)
}

func (t *Tui) Help(help [][]string) {
	var s string
	for _, line := range help {
		if line[0] == "\n" {
			s += "\n"
		} else {
			s += fmt.Sprint("[-::-][", line[0], "[][#a0a0a0::b] ", line[1], " ")
		}
	}
	t.HelpWidget.SetText(s)
}

func (t *Tui) Run() error {
	if err := t.DB.LoadData(); err != nil {
		return err
	}
	if err := t.DB.RefreshProjects(0); err != nil {
		return err
	}

	t.ProjectPane.ResetCell(t.DB.Projects)
	t.ProjectPane.Select(0, 0) // len(t.DB.Projects) is usually > 0

	t.doingPaneBlurFunc()
	t.donePaneBlurFunc()
	t.descriptionWidgetBlurFunc()

	t.pushFocus(t.ProjectPane.Box)

	if err := t.App.Run(); err != nil {
		t.App.Stop()
		return err
	}

	return nil
}
