package tui

import (
	"fmt"

	todo "github.com/1set/todotxt"
	db "github.com/apxxxxxxe/kanban.txt/internal/db"
	"github.com/pkg/errors"
	"github.com/rivo/tview"
)

type TuiInterface interface {
	SortFeed()
	SelectFeed()
	Notify(text string)
	Descript(info [][]string)
}

type Tui struct {
	Config             *db.Config
	DB                 *db.FeedDB
	App                *tview.Application
	Pages              *tview.Pages
	TodoPane           *TodoTable
	DoingPane          *TodoTable
	DonePane           *TodoTable
	DescriptionWidget  *tview.TextView
	InfoWidget         *tview.TextView
	HelpWidget         *tview.TextView
	InputWidget        *InputBox
	ColorWidget        *tview.Table
	SelectingFeeds     []*todo.Task
	LastFocusedWidget  *tview.Box
	ConfirmationStatus rune
	CurrentLeftTable   int
	IsLoading          bool
}

const (
	descriptionField          = "descPopup"
	inputField                = "InputPopup"
	colorTable                = "ColorTablePopup"
	mainPage                  = "MainPage"
	keymapPage                = "KeymapPage"
	defaultConfirmationStatus = '0'
	todoPaneTitle             = "Todo"
	doingPaneTitle            = "Doing"
	donePaneTitle             = "Done"
	descriptionWidgetTitle    = "Description"
	infoWidgetTitle           = "Info"
	colorWidgetTitle          = "Color"
)

const (
	enumTodoPane = iota
	enumDoingPane
)

var ErrImportFileNotFound = errors.Errorf(db.ImportListPath + " not found")

func NewTui() *Tui {
	tview.Styles.ContrastBackgroundColor = tview.Styles.PrimitiveBackgroundColor

	tui := &Tui{
		Config:             db.LoadOrNewConfig(),
		DB:                 db.NewDB(),
		App:                tview.NewApplication(),
		Pages:              tview.NewPages(),
		TodoPane:           &TodoTable{newTable(todoPaneTitle)},
		DoingPane:          &TodoTable{newTable(doingPaneTitle)},
		DonePane:           &TodoTable{newTable(donePaneTitle)},
		DescriptionWidget:  newTextView(descriptionWidgetTitle),
		InfoWidget:         newTextView(infoWidgetTitle),
		HelpWidget:         tview.NewTextView().SetTextAlign(1).SetDynamicColors(true),
		InputWidget:        &InputBox{InputField: newInputField(), Mode: 0},
		SelectingFeeds:     []*todo.Task{},
		LastFocusedWidget:  nil,
		ConfirmationStatus: defaultConfirmationStatus,
		CurrentLeftTable:   enumTodoPane,
		IsLoading:          false,
	}

	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(tui.TodoPane, 0, 1, false).
			AddItem(tui.DoingPane, 0, 1, false).
			AddItem(tui.DonePane, 0, 1, false),
			0, 1, false).
		AddItem(tui.DescriptionWidget, 2, 0, false).
		AddItem(tui.HelpWidget, 2, 0, false)

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

	// tui.setKeyBinding()
	// tui.setSelectionFunc()
	// tui.setFocusFunc()
	// tui.setBlurFunc()

	return tui
}

func (t *Tui) setFocus(p *tview.Box) {
	t.LastFocusedWidget = p
	t.App.SetFocus(p)
}

func (t *Tui) Descript(desc [][]string) {
	var s string
	for _, line := range desc {
		s += fmt.Sprint("[#a0a0a0::b]", line[0], "[-::-] ", line[1], "\n")
	}
	t.DescriptionWidget.SetText(s).ScrollToBeginning()
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

	if err := t.DB.LoadFeeds(); err != nil {
		return err
	}

	t.setFocus(t.TodoPane.Table.Box)

	if err := t.App.Run(); err != nil {
		t.App.Stop()
		return err
	}

	return nil
}
