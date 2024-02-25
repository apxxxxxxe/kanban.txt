package db

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/1set/todotxt"
	tsk "github.com/apxxxxxxe/kanban.txt/internal/task"
	"github.com/rivo/tview"
)

const (
	KeepDoneTaskCount = 100
	dataRoot          = "kanban"
	TodaysFeedTitle   = "Today's Articles"
	SavePrefixGroup   = "g_"
	SavePrefixFeed    = "f_"
	NoProject         = "NoProject"
	AllTasks          = "AllTasks"
	allDay            = -50
	DayCount          = 7
)

var (
	DataPath       = filepath.Join(getDataPath(), "data")
	ExportListPath = filepath.Join(getDataPath(), "list_export.txt")
	ImportPath     = filepath.Join(getDataPath(), "todo.txt")
	// ArchivePath    = filepath.Join(getDataPath(), "archive.txt")
	ConfigPath     = filepath.Join(getDataPath(), "config.json")
)

type Database struct {
	LivingTasks   todotxt.TaskList
	ArchivedTasks todotxt.TaskList
	Projects      []*Project
}

type Project struct {
	ProjectName string
	TodoTasks   todotxt.TaskList
	DoingTasks  todotxt.TaskList
	DoneTasks   todotxt.TaskList
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

func sortTaskReferences(taskList todotxt.TaskList) {
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

func (d *Database) GetTaskFromTable(t *tview.Table) (*todotxt.Task, error) {
	cell := t.GetCell(t.GetSelection())
	if cell == nil {
		return nil, errors.New("no task selected")
	}
	task, ok := cell.GetReference().(*todotxt.Task)
	if !ok {
		return nil, errors.New("the selected cell is not a *todotxt.Task")
	}
	return d.LivingTasks.GetTask(task.ID)
}

func (d *Database) SaveData() error {
  allTasks := append(d.LivingTasks, d.ArchivedTasks...)
  sortTaskList(allTasks)
	return d.saveData(allTasks, ImportPath)
}

func (d *Database) saveData(taskList todotxt.TaskList, filePath string) error {
	tasklist := todotxt.NewTaskList()
	for _, t := range taskList {
		tasklist.AddTask(&t)
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

	allTasks, err := todotxt.LoadFromPath(ImportPath)
	if err != nil {
		return err
	}

	d.LivingTasks = allTasks.Filter(todotxt.FilterNotCompleted)

	completedTasks := allTasks.Filter(todotxt.FilterCompleted)
	completedTasks.Sort(todotxt.SortCompletedDateDesc)
	for i := range completedTasks {
		if i < KeepDoneTaskCount {
			d.LivingTasks.AddTask(&completedTasks[i])
		}
	}

	d.ArchivedTasks = todotxt.NewTaskList()
	for i := range completedTasks {
		if i >= KeepDoneTaskCount {
			d.ArchivedTasks.AddTask(&completedTasks[i])
		}
	}

	return nil
}

func (d *Database) loadData(filePath string) (todotxt.TaskList, error) {
	var err error
	taskList := todotxt.TaskList{}
	taskList, err = todotxt.LoadFromPath(filePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	} else if errors.Is(err, os.ErrNotExist) {
		return taskList, nil
	}

	for i := range taskList {
		task := taskList[i]
		if task.Projects == nil || len(task.Projects) == 0 {
			task.Projects = []string{NoProject}
		}
	}

	// Remove contexts from completed tasks
	for _, task := range taskList.Filter(todotxt.FilterCompleted).Filter(todotxt.FilterByContext("doing")) {
		for j := range taskList {
			if taskList[j].ID == task.ID {
				removeContexts(&task)
			}
		}
	}

	sortTaskReferences(taskList)

	return taskList, nil
}

func (d *Database) recurrentTasks(day int) error {
	date := time.Now().AddDate(0, 0, day)
	tasks := map[string]todotxt.Task{}
	doneTasks := d.LivingTasks.Filter(todotxt.FilterCompleted)
	doneTasks.Sort(todotxt.SortCreatedDateDesc)
	for _, t := range doneTasks {
		key := tsk.GetTaskKey(t)
		if _, ok := tasks[key]; !ok {
			tasks[key] = t
		}
	}

	for i := range tasks {
		t := tasks[i]
		if _, ok := t.AdditionalTags[tsk.KeyRec]; ok {
			nextTime, err := tsk.ParseRecurrence(&t)
			if err != nil {
				return err
			}
			// nextTimeを経過しているかどうか
			if nextTime.Before(date) {
				sameTasks := d.LivingTasks.
					Filter(todotxt.FilterNotCompleted).
					Filter(filterByTodo(t.Todo)).
					Filter(todotxt.FilterByPriority(t.Priority)).
					Filter(todotxt.FilterByProject(tsk.GetProjectName(t)))
					// Filter(filterHasSameContexts(t))
				if len(sameTasks) == 0 {
					newTask := copyTask(t)
					newTask.Reopen()
					delete(newTask.AdditionalTags, tsk.KeyStartDoing)
					newTask.CreatedDate = date
					// if err := d.moveToArchive(t); err != nil {
					// 	return err
					// }
					d.LivingTasks.AddTask(&newTask)
				}
			}
		}
	}

	return nil
}

func (d *Database) moveToArchive(t todotxt.Task) error {
	if err := d.LivingTasks.RemoveTask(t); err != nil {
		return err
	}
	d.ArchivedTasks.AddTask(&t)
	return nil
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

func timeToDate(t *time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

func FilterCompareDate(date time.Time) todotxt.Predicate {
	return func(t todotxt.Task) bool {
		taskMakedDoing := time.Time{}
		if v, ok := t.AdditionalTags[tsk.KeyStartDoing]; ok {
			var err error
			taskMakedDoing, err = time.Parse(todotxt.DateLayout, v)
			if err != nil {
				panic(err)
			}
		}
		taskMakedDoing = timeToDate(&taskMakedDoing)

		taskCreated := timeToDate(&t.CreatedDate)
		taskDue := timeToDate(&t.DueDate)
		date = timeToDate(&date)

		makedDoingComp := taskMakedDoing.Compare(date)
		createdComp := taskCreated.Compare(date)
		dueComp := taskDue.Compare(date)

		// dateは作業中である or Doingされていない
		isOkMakedDoing := makedDoingComp <= 0 || taskMakedDoing.IsZero()

		// dateは作成日以降である
		isOkCreated := createdComp <= 0

		// dateは期限前である or 期限がない
		isOkDue := !t.HasDueDate() || dueComp >= 0

		return isOkMakedDoing && isOkCreated && isOkDue
	}
}

func (d *Database) RefreshProjects(day int) error {
	if err := d.recurrentTasks(day); err != nil {
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

	getTodoTasks := func(tasklist todotxt.TaskList, day int) todotxt.TaskList {
		date := time.Now().AddDate(0, 0, day)
		return tasklist.Filter(todotxt.FilterNotCompleted).Filter(todotxt.FilterNot(todotxt.FilterByContext("doing"))).Filter(FilterCompareDate(date))
	}

	getDoingTasks := func(tasklist todotxt.TaskList, day int) todotxt.TaskList {
		date := time.Now().AddDate(0, 0, day)
		return tasklist.Filter(todotxt.FilterNotCompleted).Filter(todotxt.FilterByContext("doing")).Filter(FilterCompareDate(date))
	}

	getDoneTasks := func(tasklist todotxt.TaskList, day int) todotxt.TaskList {
		date := time.Now().AddDate(0, 0, day)
		tasks := tasklist.Filter(todotxt.FilterCompleted).Filter(todotxt.FilterNot(todotxt.FilterByContext("doing"))).Filter(FilterCompareDate(date))
		tasks.Sort(todotxt.SortCompletedDateDesc)
		return tasks
	}

	list := map[string]func(todotxt.TaskList, int) todotxt.TaskList{
		todo:  getTodoTasks,
		doing: getDoingTasks,
		done:  getDoneTasks,
	}

	projectList := map[string]*Project{}
	for key, fn := range list {
		for _, task := range fn(d.LivingTasks, day) {
			projectName := tsk.GetProjectName(task)
			project, ok := projectList[projectName]
			if !ok {
				project = &Project{ProjectName: projectName}
				projectList[projectName] = project
			}

			switch key {
			case todo:
				project.TodoTasks.AddTask(&task)
			case doing:
				project.DoingTasks.AddTask(&task)
			case done:
				project.DoneTasks.AddTask(&task)
			}
		}
	}

	projects := []*Project{}
	for _, p := range projectList {
		projects = append(projects, p)
	}
	allTaskProject := &Project{
		ProjectName: AllTasks,
		TodoTasks:   getTodoTasks(d.LivingTasks, day),
		DoingTasks:  getDoingTasks(d.LivingTasks, day),
		DoneTasks:   getDoneTasks(d.LivingTasks, day),
	}
	projects = append(projects, allTaskProject)

	// sort projects
	sort.Slice(projects, func(i, j int) bool {
		// sort by project name
		// noProject is always the first
		if projects[i].ProjectName == AllTasks {
			return true
		} else if projects[j].ProjectName == AllTasks {
			return false
		} else if projects[i].ProjectName == NoProject {
			return true
		} else if projects[j].ProjectName == NoProject {
			return false
		} else {
			return projects[i].ProjectName < projects[j].ProjectName
		}
	})

	d.Projects = projects

	if err := d.SaveData(); err != nil {
		return err
	}

	return nil
}
