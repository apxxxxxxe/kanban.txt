package db

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/1set/todotxt"
	"github.com/apxxxxxxe/kanban.txt/internal/task"
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
	ArchivePath    = filepath.Join(getDataPath(), "archive.txt")
	ConfigPath     = filepath.Join(getDataPath(), "config.json")
)

type Database struct {
	LivingTasks   TaskReferences
	ArchivedTasks TaskReferences
	Projects      []*Project
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

func displayContextsWithoutDoing(t *todotxt.Task) string {
	raw := ""
	cs := t.Contexts
	sort.Strings(cs)
	for _, s := range cs {
		if s != "doing" {
			raw += s + " "
		}
	}
	return raw
}

func displayContexts(t *todotxt.Task) string {
	raw := ""
	cs := t.Contexts
	sort.Strings(cs)
	for _, s := range cs {
		raw += s + " "
	}
	return raw
}

func GetProjectName(t todotxt.Task) string {
	projects := t.Projects
	if len(projects) == 0 || projects == nil {
		panic("no project")
	}
	return projects[0]
}

func copyTask(t todotxt.Task) todotxt.Task {
	var newTask todotxt.Task
	newTask = t
	newTask.AdditionalTags = map[string]string{}
	for k, v := range t.AdditionalTags {
		newTask.AdditionalTags[k] = v
	}
	return newTask
}

func sortTaskList(taskList todotxt.TaskList) {
	sort.Slice(taskList, func(i, j int) bool {
		if taskList[i].Completed != taskList[j].Completed {
			return !taskList[i].Completed && taskList[j].Completed
		} else if taskList[i].Priority != taskList[j].Priority {
			return comparePriority(taskList[i].Priority, taskList[j].Priority)
		} else if !taskList[i].DueDate.Equal(taskList[j].DueDate) {
			return taskList[i].DueDate.Before(taskList[j].DueDate)
		} else {
			return taskList[i].Todo < taskList[j].Todo
		}
	})
}

func sortTaskReferences(taskList TaskReferences) {
	sort.Slice(taskList, func(i, j int) bool {
		if taskList[i].Completed != taskList[j].Completed {
			return !taskList[i].Completed && taskList[j].Completed
		} else if taskList[i].Priority != taskList[j].Priority {
			return comparePriority(taskList[i].Priority, taskList[j].Priority)
		} else if !taskList[i].DueDate.Equal(taskList[j].DueDate) {
			return taskList[i].DueDate.Before(taskList[j].DueDate)
		} else {
			return taskList[i].Todo < taskList[j].Todo
		}
	})
}

func (d *Database) SaveData() error {
	if err := d.saveData(d.LivingTasks, ImportPath); err != nil {
		return err
	}
	if err := d.saveData(d.ArchivedTasks, ArchivePath); err != nil {
		return err
	}
	return nil
}

func (d *Database) saveData(taskList TaskReferences, filePath string) error {
	tasklist := todotxt.NewTaskList()
	for _, t := range taskList {
		tasklist = append(tasklist, copyTask(*t))
	}

	for _, t := range tasklist {
		if t.Projects != nil && len(t.Projects) > 0 && t.Projects[0] == NoProject {
			t.Projects = nil
		}
	}

	sortTaskList(tasklist)

	fp, err := os.Create(filePath)
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

func comparePriority(p1, p2 string) bool {
	if p1 == "" {
		return false
	} else if p2 == "" {
		return true
	} else {
		return p1 < p2
	}
}

func (d *Database) LoadData() error {
	var err error

	d.LivingTasks, err = d.loadData(ImportPath)
	if err != nil {
		return err
	}

	d.ArchivedTasks, err = d.loadData(ArchivePath)
	if err != nil {
		return err
	}

	d.Projects = []*Project{}
	return nil
}

func (d *Database) loadData(filePath string) (TaskReferences, error) {
	var err error
	taskList := TaskReferences{}
	tmpList := todotxt.TaskList{}
	tmpList, err = todotxt.LoadFromPath(filePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	} else if errors.Is(err, os.ErrNotExist) {
		return taskList, nil
	}

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

	sortTaskReferences(taskList)

	return taskList, nil
}

func (d *Database) recurrentTasks() error {
	now := time.Now()
	tasks := map[string]*todotxt.Task{}
	doneTasks := *d.LivingTasks.Filter(todotxt.FilterCompleted)
	doneTasks.Sort(todotxt.SortCreatedDateDesc)
	for _, t := range doneTasks {
		key := t.Todo + t.Priority + GetProjectName(*t)
		if _, ok := tasks[key]; !ok {
			tasks[key] = t
		}
	}

	for _, t := range tasks {
		if _, ok := t.AdditionalTags[task.KeyRec]; ok {
			nextTime, err := task.ParseRecurrence(t)
			if err != nil {
				return err
			}
			if nextTime.Before(now) {
				sameTasks := d.LivingTasks.
					Filter(todotxt.FilterNotCompleted).
					Filter(filterByTodo(t.Todo)).
					Filter(todotxt.FilterByPriority(t.Priority)).
					Filter(todotxt.FilterByProject(GetProjectName(*t))).
					Filter(filterHasSameContexts(*t))
				if len(*sameTasks) == 0 {
					newTask := copyTask(*t)
					newTask.Reopen()
					newTask.CreatedDate = now
					d.LivingTasks.AddTask(&newTask)
					d.moveToArchive(t)
				}
			}
		}
	}

	return nil
}

func (d *Database) moveToArchive(t *todotxt.Task) {
	d.LivingTasks.RemoveTask(t)
	d.ArchivedTasks.AddTask(t)
}

func filterByTodo(todo string) todotxt.Predicate {
	return func(t todotxt.Task) bool {
		return t.Todo == todo
	}
}

func filterHasSameContexts(a todotxt.Task) todotxt.Predicate {
	return func(t todotxt.Task) bool {
		return displayContextsWithoutDoing(&a) == displayContextsWithoutDoing(&t)
	}
}

func (d *Database) RefreshProjects() error {
	if err := d.recurrentTasks(); err != nil {
		return err
	}

	sortTaskReferences(d.LivingTasks)

	d.Projects = []*Project{}

	if len(d.LivingTasks) == 0 {
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
		tasks := tasklist.Filter(todotxt.FilterCompleted).Filter(todotxt.FilterNot(todotxt.FilterByContext("doing")))
		tasks.Sort(todotxt.SortCompletedDateDesc)
		return *tasks
	}

	list := map[string]func(TaskReferences) TaskReferences{
		todo:  getTodoTasks,
		doing: getDoingTasks,
		done:  getDoneTasks,
	}

	projectList := map[string]*Project{}
	for key, fn := range list {
		for _, task := range fn(d.LivingTasks) {
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
	allTasks.TodoTasks = getTodoTasks(d.LivingTasks)
	allTasks.DoingTasks = getDoingTasks(d.LivingTasks)
	allTasks.DoneTasks = getDoneTasks(d.LivingTasks)
	d.Projects = append([]*Project{allTasks}, d.Projects...)

	if err := d.SaveData(); err != nil {
		return err
	}

	return nil
}
