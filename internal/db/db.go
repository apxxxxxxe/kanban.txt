package db

import (
	"os"
	"path/filepath"

	"github.com/1set/todotxt"
)

const (
	dataRoot        = "kanban"
	TodaysFeedTitle = "Today's Articles"
	SavePrefixGroup = "g_"
	SavePrefixFeed  = "f_"
	noProject       = "NoProject"
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
	TodoTasks   todotxt.TaskList
	DoingTasks  todotxt.TaskList
	DoneTasks   todotxt.TaskList
}

func getDataPath() string {
	// configDir, _ := os.UserConfigDir()
	// return filepath.Join(configDir, dataRoot)
	wd, _ := os.Getwd()
	return wd
}

func removeContexts(t todotxt.Task) todotxt.Task {
	raw := ""
	for _, s := range t.Segments() {
		if s.Type != todotxt.SegmentContext {
			raw += s.Display + " "
		}
	}
	parsed, _ := todotxt.ParseTask(raw)
	return *parsed
}

func GetProjectName(t todotxt.Task) string {
	projects := t.Projects
	if len(projects) == 0 || projects == nil {
		return noProject
	}
	return projects[0]
}

func (d *Database) Reset() {
	d.Projects = []*Project{
		{
			ProjectName: noProject,
			TodoTasks:   todotxt.NewTaskList(),
		},
	}
}

func (d *Database) SaveFeeds() error {
  tasklist := todotxt.NewTaskList()

  for _, project := range d.Projects {
    for _, task := range project.TodoTasks {
      tasklist = append(tasklist, task)
    }
    for _, task := range project.DoingTasks {
      tasklist = append(tasklist, task)
    }
    for _, task := range project.DoneTasks {
      tasklist = append(tasklist, task)
    }
  }

  if err := tasklist.Sort(todotxt.SortPriorityAsc, todotxt.SortDueDateAsc); err != nil {
    return err
  }

  fp, err := os.Create(ImportPath)
  if err != nil {
    return err
  }
  defer fp.Close()

  for _, task := range tasklist {
    _, err := fp.WriteString(task.String() + "\n")
    if err != nil {
      return err
    }
  }

  return nil
}

func (d *Database) LoadFeeds() error {
	tasklist, err := todotxt.LoadFromPath(ImportPath)
	if err != nil {
		return err
	}

	d.Reset()

	// Remove contexts from completed tasks
	for _, task := range tasklist.Filter(todotxt.FilterCompleted).Filter(todotxt.FilterByContext("doing")) {
		for j := range tasklist {
			if tasklist[j].ID == task.ID {
				tasklist[j] = removeContexts(task)
			}
		}
	}

	err = tasklist.Sort(todotxt.SortPriorityAsc, todotxt.SortDueDateAsc)
	if err != nil {
		return err
	}

  const (
    todo = "todo"
    doing = "doing"
    done = "done"
  )

  getTodoTasks := func(tasklist todotxt.TaskList) todotxt.TaskList {
    return tasklist.Filter(todotxt.FilterNotCompleted).Filter(todotxt.FilterNot(todotxt.FilterByContext("doing")))
  }

  getDoingTasks := func(tasklist todotxt.TaskList) todotxt.TaskList {
    return tasklist.Filter(todotxt.FilterNotCompleted).Filter(todotxt.FilterByContext("doing"))
  }

  getDoneTasks := func(tasklist todotxt.TaskList) todotxt.TaskList {
    return tasklist.Filter(todotxt.FilterCompleted).Filter(todotxt.FilterNot(todotxt.FilterByContext("doing")))
  }

  list := map[string]func(todotxt.TaskList) todotxt.TaskList{
    todo: getTodoTasks,
    doing: getDoingTasks,
    done: getDoneTasks,
  }

	projectList := map[string]*Project{}
  for key, fn := range list {
    for _, task := range fn(tasklist) {
      projectName := GetProjectName(task)
      project, ok := projectList[projectName]
      if !ok {
        d.Projects = append(d.Projects, &Project{ProjectName: projectName})
        projectList[projectName] = d.Projects[len(d.Projects)-1]
        project = d.Projects[len(d.Projects)-1]
      }

      switch key {
      case todo:
        project.TodoTasks = append(project.TodoTasks, task)
      case doing:
        project.DoingTasks = append(project.DoingTasks, task)
      case done:
        project.DoneTasks = append(project.DoneTasks, task)
      }
    }
  }

	return nil
}
