package db

import (
	"os"
	"path/filepath"

	todo "github.com/1set/todotxt"
)

const (
	dataRoot        = "fd"
	TodaysFeedTitle = "Today's Articles"
	SavePrefixGroup = "g_"
	SavePrefixFeed  = "f_"
)

var (
	DataPath       = filepath.Join(getDataPath(), "data")
	ExportListPath = filepath.Join(getDataPath(), "list_export.txt")
	ImportPath     = filepath.Join(getDataPath(), "todo.txt")
	ConfigPath     = filepath.Join(getDataPath(), "config.json")
)

type Database struct {
	TodoTasks  todo.TaskList
	DoingTasks todo.TaskList
	DoneTasks  todo.TaskList
}

func NewDB() *Database {
	db := &Database{
		TodoTasks:  todo.TaskList{},
		DoingTasks: todo.TaskList{},
		DoneTasks:  todo.TaskList{},
	}
	return db
}

func getDataPath() string {
	configDir, _ := os.Getwd()
	// configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, dataRoot)
}

func RemoveContexts(t todo.Task) todo.Task {
	raw := ""
	for _, s := range t.Segments() {
		if s.Type != todo.SegmentContext {
			raw += s.Display + " "
		}
	}
	parsed, _ := todo.ParseTask(raw)
	return *parsed
}

func (d *Database) LoadFeeds() error {
	tasklist, err := todo.LoadFromPath(ImportPath)
	if err != nil {
		return err
	}

	// Remove contexts from completed tasks
	for _, task := range tasklist.Filter(todo.FilterCompleted).Filter(todo.FilterByContext("doing")) {
		for j := range tasklist {
			if tasklist[j].ID == task.ID {
				tasklist[j] = RemoveContexts(task)
			}
		}
	}

	err = tasklist.Sort(todo.SortPriorityAsc, todo.SortDueDateAsc)
	if err != nil {
		return err
	}

	d.TodoTasks = tasklist.Filter(todo.FilterNotCompleted).Filter(todo.FilterNot(todo.FilterByContext("doing")))
	d.DoingTasks = tasklist.Filter(todo.FilterNotCompleted).Filter(todo.FilterByContext("doing"))
	d.DoneTasks = tasklist.Filter(todo.FilterCompleted).Filter(todo.FilterNot(todo.FilterByContext("doing")))

	return nil
}
