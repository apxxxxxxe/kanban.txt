package db

import (
	"encoding/json"
	"errors"
	"github.com/1set/todotxt"
	tsk "github.com/apxxxxxxe/kanban.txt/internal/task"
	"os"
	"path/filepath"
	"sort"
	"time"
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
	ArchivePath    = filepath.Join(getDataPath(), "archive.json")
	ImportPath     = filepath.Join(getDataPath(), "todo.txt")
	ConfigPath     = filepath.Join(getDataPath(), "config.json")
)

type Database struct {
	LivingTasks   TaskReferences
	HiddenTasks   TaskReferences
	ArchivedTasks []string
	Projects      []*Project
}

type Archive struct {
	ArchivedTasks []string `json:"archived_tasks"`
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

func copyTask(t todotxt.Task) todotxt.Task {
	var newTask todotxt.Task
	newTask = t
	newTask.AdditionalTags = map[string]string{}
	for k, v := range t.AdditionalTags {
		newTask.AdditionalTags[k] = v
	}
	return newTask
}

func sortTaskReferences(taskList TaskReferences) {
	sort.Slice(taskList, func(i, j int) bool {
		return taskList[i].String() < taskList[j].String()
	})
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
	return d.saveData(append(d.LivingTasks, d.HiddenTasks...), ImportPath)
}

func (d *Database) saveData(taskList TaskReferences, filePath string) error {
	tasklist := TaskReferences{}
	for _, t := range taskList {
		ct := copyTask(*t)
		tasklist = append(tasklist, &ct)
	}

	for _, t := range tasklist {
		if t.Projects != nil && len(t.Projects) > 0 && t.Projects[0] == NoProject {
			t.Projects = nil
		}
	}

	sortTaskReferences(tasklist)

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

	b, err := json.MarshalIndent(Archive{ArchivedTasks: d.ArchivedTasks}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(ArchivePath, b, 0644); err != nil {
		return err
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

	allTasks, err := d.loadData(ImportPath)
	if err != nil {
		return err
	}

	allTaskMap := makeTaskMap(allTasks)
	idArray := []int{}
	for k := range allTaskMap {
		idArray = append(idArray, allTaskMap[k].ID)
	}
	sort.Ints(idArray)

	d.LivingTasks = *allTasks.Filter(FilterMapContains(idArray))

	d.HiddenTasks = *allTasks.Filter(todotxt.FilterNot(FilterMapContains(idArray)))

	var archive Archive
	b, err := os.ReadFile(ArchivePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		archive = Archive{}
	} else {
		if err := json.Unmarshal(b, &archive); err != nil {
			return err
		}
	}
	d.ArchivedTasks = archive.ArchivedTasks

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
		task := &tmpList[i]
		if task.Projects == nil || len(task.Projects) == 0 {
			task.Projects = []string{NoProject}
		}
		taskList = append(taskList, task)
	}

	//Remove contexts from completed tasks
	filteredTasks := tmpList.Filter(todotxt.FilterCompleted).Filter(todotxt.FilterByContext("doing"))
	for i := range filteredTasks {
		task := filteredTasks[i]
		for j := range taskList {
			if taskList[j].ID == task.ID {
				removeContexts(&task)
			}
		}
	}

	sortTaskReferences(taskList)

	return taskList, nil
}

func makeTaskMap(taskList TaskReferences) map[string]*todotxt.Task {
	sort.Slice(taskList, func(i, j int) bool {
		return taskList[i].CompletedDate.After(taskList[j].CompletedDate) ||
			taskList[i].CreatedDate.After(taskList[j].CreatedDate)
	})
	taskMap := map[string]*todotxt.Task{}
	for i := range taskList {
		t := taskList[i]
		key := tsk.GetTaskKey(*t)
		if _, ok := taskMap[key]; !ok {
			taskMap[key] = t
		}
	}
	return taskMap
}

func (d *Database) ArchiveTask(t *todotxt.Task) {
	v, ok := t.AdditionalTags[tsk.KeyRecID]
	if !ok {
		panic("recurrence id not found: " + t.String())
	}
	d.ArchivedTasks = append(d.ArchivedTasks, v)
}

func (d *Database) recurrentTasks(day int) error {
	date := time.Now().AddDate(0, 0, day)
	doneTasks := *d.LivingTasks.Filter(todotxt.FilterCompleted)
	sort.Slice(doneTasks, func(i, j int) bool {
		return doneTasks[i].CreatedDate.Before(doneTasks[j].CreatedDate)
	})

	taskMap := makeTaskMap(doneTasks)

	tasks := TaskReferences{}
	for _, v := range taskMap {
		tasks = append(tasks, v)
	}
	sortTaskReferences(tasks)

	for i := range tasks {
		t := tasks[i]
		if _, ok := t.AdditionalTags[tsk.KeyRec]; ok {
			nextTime, err := tsk.ParseRecurrence(t)
			if err != nil {
				return err
			}
			// nextTimeを経過しているかどうか
			if nextTime.Before(date) {
				newTask := copyTask(*t)
				newTask.Reopen()
				delete(newTask.AdditionalTags, tsk.KeyStartDoing)
				newTask.CreatedDate = date
				d.LivingTasks.AddTask(&newTask)
			}
		}
	}

	return nil
}

func (d *Database) moveToArchive(t *todotxt.Task) error {
	d.LivingTasks.RemoveTask(t)
	d.HiddenTasks.AddTask(t)
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

func FilterArchivedTasks(archivedTasks []string) todotxt.Predicate {
	return func(t todotxt.Task) bool {
		for _, key := range archivedTasks {
			if v, ok := t.AdditionalTags[tsk.KeyRecID]; ok && v == key {
				return true
			}
		}
		return false
	}
}

func FilterMapContains(idArray []int) todotxt.Predicate {
	return func(t todotxt.Task) bool {
		for _, id := range idArray {
			if t.ID == id {
				return true
			}
		}
		return false
	}
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

	getTodoTasks := func(tasklist TaskReferences, day int) TaskReferences {
		date := time.Now().AddDate(0, 0, day)
		return *tasklist.Filter(todotxt.FilterNotCompleted).
			Filter(todotxt.FilterNot(todotxt.FilterByContext("doing"))).
			Filter(FilterCompareDate(date)).
			Filter(todotxt.FilterNot(FilterArchivedTasks(d.ArchivedTasks)))
	}

	getDoingTasks := func(tasklist TaskReferences, day int) TaskReferences {
		date := time.Now().AddDate(0, 0, day)
		return *tasklist.Filter(todotxt.FilterNotCompleted).
			Filter(todotxt.FilterByContext("doing")).
			Filter(FilterCompareDate(date)).
			Filter(todotxt.FilterNot(FilterArchivedTasks(d.ArchivedTasks)))
	}

	getDoneTasks := func(tasklist TaskReferences, day int) TaskReferences {
		date := time.Now().AddDate(0, 0, day)
		tasks := *tasklist.Filter(todotxt.FilterCompleted).
			Filter(todotxt.FilterNot(todotxt.FilterByContext("doing"))).
			Filter(FilterCompareDate(date)).
			Filter(todotxt.FilterNot(FilterArchivedTasks(d.ArchivedTasks)))
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].CompletedDate.After(tasks[j].CompletedDate)
		})
		return tasks
	}

	list := map[string]func(TaskReferences, int) TaskReferences{
		todo:  getTodoTasks,
		doing: getDoingTasks,
		done:  getDoneTasks,
	}

	projectList := map[string]*Project{}
	for key, fn := range list {
		for _, task := range fn(d.LivingTasks, day) {
			projectName := tsk.GetProjectName(*task)
			project, ok := projectList[projectName]
			if !ok {
				project = &Project{ProjectName: projectName}
				projectList[projectName] = project
			}

			switch key {
			case todo:
				project.TodoTasks.AddTask(task)
			case doing:
				project.DoingTasks.AddTask(task)
			case done:
				project.DoneTasks.AddTask(task)
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
