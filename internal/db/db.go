package db

import (
	"os"
	"path/filepath"

	todo "github.com/1set/todotxt"
)

const (
	dataRoot        = "kanban"
	TodaysFeedTitle = "Today's Articles"
	SavePrefixGroup = "g_"
	SavePrefixFeed  = "f_"
	noProject       = "NoName"
)

var (
	DataPath       = filepath.Join(getDataPath(), "data")
	ExportListPath = filepath.Join(getDataPath(), "list_export.txt")
	ImportPath     = filepath.Join(getDataPath(), "todo.txt")
	ConfigPath     = filepath.Join(getDataPath(), "config.json")
)

type Database struct {
	Projects []*Project
}

type Project struct {
	ProjectName string
	TodoTasks   todo.TaskList
	DoingTasks  todo.TaskList
	DoneTasks   todo.TaskList
}

func getDataPath() string {
	// configDir, _ := os.UserConfigDir()
	// return filepath.Join(configDir, dataRoot)
	wd, _ := os.Getwd()
	return wd
}

func removeContexts(t todo.Task) todo.Task {
	raw := ""
	for _, s := range t.Segments() {
		if s.Type != todo.SegmentContext {
			raw += s.Display + " "
		}
	}
	parsed, _ := todo.ParseTask(raw)
	return *parsed
}

func GetProjectName(t todo.Task) string {
	projects := t.Projects
	if len(projects) == 0 || projects == nil {
		return noProject
	}
	return projects[0]
}

func (d *Database) GetProjectByName(name string) *Project {
  for _, p := range d.Projects {
    if p.ProjectName == name {
      return p
    }
  }
  return nil
}

func (d *Database) Reset() {
	d.Projects = []*Project{
		{
			ProjectName: noProject,
			TodoTasks:   todo.NewTaskList(),
		},
	}
}

func (d *Database) LoadFeeds() error {
	tasklist, err := todo.LoadFromPath(ImportPath)
	if err != nil {
		return err
	}

	d.Reset()

	// Remove contexts from completed tasks
	for _, task := range tasklist.Filter(todo.FilterCompleted).Filter(todo.FilterByContext("doing")) {
		for j := range tasklist {
			if tasklist[j].ID == task.ID {
				tasklist[j] = removeContexts(task)
			}
		}
	}

	err = tasklist.Sort(todo.SortPriorityAsc, todo.SortDueDateAsc)
	if err != nil {
		return err
	}

	projectList := map[string]int{}
	projectCount := 0

	for _, todo := range tasklist.Filter(todo.FilterNotCompleted).Filter(todo.FilterNot(todo.FilterByContext("doing"))) {
		project := GetProjectName(todo)

		if _, ok := projectList[project]; !ok {
			projectList[project] = projectCount
			projectCount++
			d.Projects = append(d.Projects, &Project{ProjectName: project})
		}

		d.Projects[projectList[project]].TodoTasks = append(d.Projects[projectList[project]].TodoTasks, todo)
	}

	for _, todo := range tasklist.Filter(todo.FilterNotCompleted).Filter(todo.FilterByContext("doing")) {
		project := GetProjectName(todo)

		if _, ok := projectList[project]; !ok {
			projectList[project] = projectCount
			projectCount++
			d.Projects = append(d.Projects, &Project{ProjectName: project})
		}

		d.Projects[projectList[project]].DoingTasks = append(d.Projects[projectList[project]].DoingTasks, todo)
	}

	for _, todo := range tasklist.Filter(todo.FilterCompleted).Filter(todo.FilterNot(todo.FilterByContext("doing"))) {
		project := GetProjectName(todo)

		if _, ok := projectList[project]; !ok {
			projectList[project] = projectCount
			projectCount++
			d.Projects = append(d.Projects, &Project{ProjectName: project})
		}

		d.Projects[projectList[project]].DoneTasks = append(d.Projects[projectList[project]].DoneTasks, todo)
	}

	return nil
}
