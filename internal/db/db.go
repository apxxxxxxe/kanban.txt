package db

import (
	"os"
	"path/filepath"

	todo "github.com/1set/todotxt"
	"github.com/apxxxxxxe/kanban.txt/pkg/util"
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
	ImportListPath = filepath.Join(getDataPath(), "list.txt")
	ConfigPath     = filepath.Join(getDataPath(), "config.json")
)

type FeedDB struct {
	TodoTasks  []*todo.Task
	DoingTasks []*todo.Task
	DoneTasks  []*todo.Task
}

func NewDB() *FeedDB {
	db := &FeedDB{
		TodoTasks:  []*todo.Task{},
		DoingTasks: []*todo.Task{},
		DoneTasks:  []*todo.Task{},
	}
	return db
}

func getDataPath() string {
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, dataRoot)
}

func (d *FeedDB) LoadFeeds() error {
	if !util.IsDir(DataPath) {
		if err := os.MkdirAll(DataPath, 0755); err != nil {
			return err
		}
	}

	return nil
}
