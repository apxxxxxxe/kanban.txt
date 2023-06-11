package db

import (
	"errors"
	"os"
	"path/filepath"
	"sort"

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
	WholeTasks TaskReferences
	Projects   []*Project
}

type Project struct {
	ProjectName string
	TodoTasks   TaskReferences
	DoingTasks  TaskReferences
	DoneTasks   TaskReferences
}

func getDataPath() string {
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, dataRoot)
}

func removeContexts(t *todotxt.Task) {
	raw := ""
	for _, s := range t.Segments() {
		if s.Type != todotxt.SegmentContext {
			raw += s.Display + " "
		}
	}
	t, _ = todotxt.ParseTask(raw)
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
			TodoTasks:   TaskReferences{},
		},
	}
}

func (d *Database) SaveData() error {
	tasklist := todotxt.NewTaskList()
	for i := range d.WholeTasks {
		tasklist = append(tasklist, *d.WholeTasks[i])
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

func (d *Database) LoadData() error {
	var err error
	tmpList := todotxt.TaskList{}
	tmpList, err = todotxt.LoadFromPath(ImportPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	taskList := TaskReferences{}
	for i := range tmpList {
		taskList = append(taskList, &tmpList[i])
	}

	// Remove contexts from completed tasks
	for _, task := range tmpList.Filter(todotxt.FilterCompleted).Filter(todotxt.FilterByContext("doing")) {
		for j := range taskList {
			if taskList[j].ID == task.ID {
				removeContexts(&task)
			}
		}
	}

	if err = taskList.Sort(todotxt.SortPriorityAsc, todotxt.SortDueDateAsc); err != nil {
		return err
	}

	d.Reset()
	d.WholeTasks = taskList

	return nil
}

func (d *Database) RefreshProjects() error {
	d.Projects = []*Project{{ProjectName: noProject}}

	if len(d.WholeTasks) == 0 {
		return nil
	}

	const (
		todo  = "todo"
		doing = "doing"
		done  = "done"
	)

	getTodoTasks := func(tasklist TaskReferences) TaskReferences {
		return *tasklist.Filter(todotxt.FilterNotCompleted).Filter(todotxt.FilterNot(todotxt.FilterByContext("doing")))
	}

	getDoingTasks := func(tasklist TaskReferences) TaskReferences {
		return *tasklist.Filter(todotxt.FilterNotCompleted).Filter(todotxt.FilterByContext("doing"))
	}

	getDoneTasks := func(tasklist TaskReferences) TaskReferences {
		return *tasklist.Filter(todotxt.FilterCompleted).Filter(todotxt.FilterNot(todotxt.FilterByContext("doing")))
	}

	list := map[string]func(TaskReferences) TaskReferences{
		todo:  getTodoTasks,
		doing: getDoingTasks,
		done:  getDoneTasks,
	}

	projectList := map[string]*Project{}
	for key, fn := range list {
		for _, task := range fn(d.WholeTasks) {
			projectName := GetProjectName(*task)
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

	// sort projects(tasks are already sorted)
	sort.Slice(d.Projects, func(i, j int) bool {
		return d.Projects[i].ProjectName < d.Projects[j].ProjectName
	})

	if err := d.SaveData(); err != nil {
		return err
	}

	return nil
}

func ToTodo(task *todotxt.Task) {
	for i, c := range task.Contexts {
		if c == "doing" {
			task.Contexts = append(task.Contexts[:i], task.Contexts[i+1:]...)
			break
		}
	}
	task.Reopen()
}

func ToDoing(task *todotxt.Task) {
	hasDoing := false
	for _, c := range task.Contexts {
		if c == "doing" {
			hasDoing = true
			break
		}
	}
	if !hasDoing {
		task.Contexts = append(task.Contexts, "doing")
	}
	task.Reopen()
}

func ToDone(task *todotxt.Task) {
	for i, c := range task.Contexts {
		if c == "doing" {
			task.Contexts = append(task.Contexts[:i], task.Contexts[i+1:]...)
			break
		}
	}
	task.Complete()
}
