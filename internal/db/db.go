package db

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/1set/todotxt"
	tsk "github.com/apxxxxxxe/kanban.txt/internal/task"
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
		if taskList[i].Completed != taskList[j].Completed {
			return !taskList[i].Completed && taskList[j].Completed
		} else if taskList[i].Priority != taskList[j].Priority {
			return comparePriority(taskList[i].Priority, taskList[j].Priority)
		} else if !taskList[i].DueDate.Equal(taskList[j].DueDate) {
			return taskList[i].DueDate.Before(taskList[j].DueDate)
		} else {
			return taskList[i].String() < taskList[j].String()
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

func devideTasks(tasks TaskReferences) (TaskReferences, TaskReferences) {
	allTaskMap := makeTaskMap(tasks, tsk.GetTaskKey)
	idArray := []int{}
	for k := range allTaskMap {
		idArray = append(idArray, allTaskMap[k].ID)
	}
	sort.Ints(idArray)

	return *tasks.Filter(filterMapContains(idArray)),
		*tasks.Filter(todotxt.FilterNot(filterMapContains(idArray)))
}

func (d *Database) LoadData() error {
	var err error

	allTasks, err := d.loadData(ImportPath)
	if err != nil {
		return err
	}

	d.LivingTasks, d.HiddenTasks = devideTasks(allTasks)

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

func makeTaskMap(taskList TaskReferences, keyFunc func(todotxt.Task) string) map[string]*todotxt.Task {
	sort.Slice(taskList, func(i, j int) bool {
		if taskList[i].Completed != taskList[j].Completed {
			return !taskList[i].Completed && taskList[j].Completed
		} else {
			return taskList[i].CreatedDate.After(taskList[j].CreatedDate)
		}
	})
	taskMap := map[string]*todotxt.Task{}
	for i := range taskList {
		t := taskList[i]
		key := keyFunc(*t)
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

func (d *Database) recurrentTasks(tasks *TaskReferences, day int) error {
	date := time.Now().AddDate(0, 0, day)
	sort.Slice((*tasks), func(i, j int) bool {
		if (*tasks)[i].Completed != (*tasks)[j].Completed {
			return (*tasks)[i].CompletedDate.After((*tasks)[j].CompletedDate)
		} else {
			return (*tasks)[i].CreatedDate.After((*tasks)[j].CreatedDate)
		}
	})

	taskMap := makeTaskMap(*tasks.Filter(filterHasRecID()), func(t todotxt.Task) string {
		return t.AdditionalTags[tsk.KeyRecID]
	})

	recurrenceCandidates := TaskReferences{}
	for _, v := range taskMap {
		if v.Completed {
			// 最新が完了済みのもののみ繰り返し判定を行う
			recurrenceCandidates = append(recurrenceCandidates, v)
		}
	}
	sortTaskReferences(recurrenceCandidates)

	for i := range recurrenceCandidates {
		t := recurrenceCandidates[i]
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
				tasks.AddTask(&newTask)
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

func filterHasRec() todotxt.Predicate {
	return func(t todotxt.Task) bool {
		_, ok := t.AdditionalTags[tsk.KeyRec]
		return ok
	}
}

func filterHasRecID() todotxt.Predicate {
	return func(t todotxt.Task) bool {
		_, ok := t.AdditionalTags[tsk.KeyRecID]
		return ok
	}
}

func timeToDate(t *time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}

func filterArchivedTasks(archivedTasks []string) todotxt.Predicate {
	return func(t todotxt.Task) bool {
		for _, key := range archivedTasks {
			if v, ok := t.AdditionalTags[tsk.KeyRecID]; ok && v == key {
				return true
			}
		}
		return false
	}
}

func filterMapContains(idArray []int) todotxt.Predicate {
	return func(t todotxt.Task) bool {
		for _, id := range idArray {
			if t.ID == id {
				return true
			}
		}
		return false
	}
}

func filterCompareDate(date time.Time) todotxt.Predicate {
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

func uniqueTaskReferences(tasks TaskReferences) TaskReferences {
	unique := TaskReferences{}
	taskString := ""
	for i := range tasks {
		t := tasks[i]
		if t.String() != taskString {
			unique = append(unique, t)
		}
		taskString = t.String()
	}
	return unique
}

// 1. allTasksにLivingとHiddenを統合
// 2. RecIDを持たないタスクにRecIDを付与
// 3. 繰り返しタスクを生成
// 4. プロジェクトごとにタスクを分類
// 5. タスクをプロジェクトごとに分類
// 6. タスクをLivingとHiddenに分類
// 7. データを保存

func (d *Database) RefreshProjects(day int) error {
	allTasks := append(d.LivingTasks, d.HiddenTasks...)
	sortTaskReferences(allTasks)
	recIDMap := map[string]string{}
	tasksHasRecID := *allTasks.Filter(filterHasRecID())
	for i := range tasksHasRecID {
		t := tasksHasRecID[i]
		recIDMap[tsk.GetTaskKey(*t)] = t.AdditionalTags[tsk.KeyRecID]
	}
	tasksNotHasRecID := *allTasks.Filter(todotxt.FilterNot(filterHasRecID())).Filter(filterHasRec())
	for i := range tasksNotHasRecID {
		t := tasksNotHasRecID[i]
		// RecIDを持たないことはFilterNot(filterHasRecID())で確認済み
		if recID, ok := recIDMap[tsk.GetTaskKey(*t)]; ok {
			t.AdditionalTags[tsk.KeyRecID] = recID
		}
	}

	if err := d.recurrentTasks(&allTasks, day); err != nil {
		return err
	}

	sortTaskReferences(allTasks)

	d.Projects = []*Project{}

	if len(allTasks) == 0 {
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
			Filter(filterCompareDate(date)).
			Filter(todotxt.FilterNot(filterArchivedTasks(d.ArchivedTasks)))
	}

	getDoingTasks := func(tasklist TaskReferences, day int) TaskReferences {
		date := time.Now().AddDate(0, 0, day)
		return *tasklist.Filter(todotxt.FilterNotCompleted).
			Filter(todotxt.FilterByContext("doing")).
			Filter(filterCompareDate(date)).
			Filter(todotxt.FilterNot(filterArchivedTasks(d.ArchivedTasks)))
	}

	getDoneTasks := func(tasklist TaskReferences, day int) TaskReferences {
		date := time.Now().AddDate(0, 0, day)
		tasks := *tasklist.Filter(todotxt.FilterCompleted).
			Filter(todotxt.FilterNot(todotxt.FilterByContext("doing"))).
			Filter(filterCompareDate(date)).
			Filter(todotxt.FilterNot(filterArchivedTasks(d.ArchivedTasks)))
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
		for _, task := range fn(allTasks, day) {
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
		TodoTasks:   getTodoTasks(allTasks, day),
		DoingTasks:  getDoingTasks(allTasks, day),
		DoneTasks:   getDoneTasks(allTasks, day),
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

	d.LivingTasks, d.HiddenTasks = devideTasks(uniqueTaskReferences(allTasks))

	if err := d.SaveData(); err != nil {
		return err
	}

	return nil
}
