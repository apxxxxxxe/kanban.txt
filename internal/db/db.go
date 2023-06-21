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
	NoProject       = "NoProject"
	AllTasks        = "AllTasks"
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
		panic("no project")
	}
	return projects[0]
}

func (d *Database) SaveData() error {
	tasklist := todotxt.NewTaskList()
	for i := range d.WholeTasks {
		task := *d.WholeTasks[i]
		if task.Projects != nil && len(task.Projects) > 0 && task.Projects[0] == NoProject {
			task.Projects = nil
		}
		tasklist = append(tasklist, task)
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
		task := tmpList[i]
		if task.Projects == nil || len(task.Projects) == 0 {
			task.Projects = []string{NoProject}
		}
		taskList = append(taskList, &task)
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

	d.Projects = []*Project{}
	d.WholeTasks = taskList

	return nil
}

func (d *Database) RefreshProjects() error {
	// sort whole tasks
	if err := d.WholeTasks.Sort(todotxt.SortPriorityAsc, todotxt.SortDueDateAsc, todotxt.SortTodoTextAsc); err != nil {
		return err
	}

	d.Projects = []*Project{}

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
				project = &Project{ProjectName: projectName}
				d.Projects = append(d.Projects, project)
				projectList[projectName] = project
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

	// sort projects
	sort.Slice(d.Projects, func(i, j int) bool {
		// sort by project name
		// noProject is always the first
		if d.Projects[i].ProjectName == NoProject {
			return true
		} else if d.Projects[j].ProjectName == NoProject {
			return false
		} else {
			return d.Projects[i].ProjectName < d.Projects[j].ProjectName
		}
	})

	// add AllTasks to the first
	allTasks := &Project{ProjectName: AllTasks}
	allTasks.TodoTasks = getTodoTasks(d.WholeTasks)
	allTasks.DoingTasks = getDoingTasks(d.WholeTasks)
	allTasks.DoneTasks = getDoneTasks(d.WholeTasks)
	d.Projects = append([]*Project{allTasks}, d.Projects...)

	if err := d.SaveData(); err != nil {
		return err
	}

	return nil
}
