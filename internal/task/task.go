package task

import (
	"errors"
	"github.com/1set/todotxt"
	"strconv"
	"strings"
	"time"
)

const (
	KeyRec        = "rec"   // 繰り返し情報
	KeyRecID      = "recid" // 繰り返し情報のID
	KeyNote       = "note"  // 備考
	KeyStartDoing = "doing" // Doingにした日時
)

func GetProjectName(t todotxt.Task) string {
	projects := t.Projects
	if len(projects) == 0 || projects == nil {
		panic("no project")
	}
	return projects[0]
}

func GetTaskKey(t todotxt.Task) string {
	if recID, ok := t.AdditionalTags[KeyRecID]; ok {
		return recID
	}
	return t.Todo + t.Priority + GetProjectName(t)
}

func ReplaceInvalidTag(field string) string {
	validTags := []string{
		KeyRec,
		KeyNote,
	}
	if !strings.Contains(field, ":") {
		return field
	}
	isInvalid := true
	for _, tag := range validTags {
		if strings.HasPrefix(field, tag+":") {
			isInvalid = false
			break
		}
	}
	if isInvalid {
		return strings.Replace(field, ":", "\\:", 1)
	}
	return field
}

func ToTodo(task *todotxt.Task) {
	for i, c := range task.Contexts {
		if c == "doing" {
			task.Contexts = append(task.Contexts[:i], task.Contexts[i+1:]...)
			break
		}
	}

	if task.AdditionalTags != nil {
		delete(task.AdditionalTags, KeyStartDoing)
	}

	task.Reopen()
}

func ToDoing(task *todotxt.Task, date time.Time) {
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

	if task.AdditionalTags == nil {
		task.AdditionalTags = make(map[string]string)
	}
	task.AdditionalTags[KeyStartDoing] = date.Format(todotxt.DateLayout)

	task.Reopen()
}

func ToDone(task *todotxt.Task, date time.Time) {
	for i, c := range task.Contexts {
		if c == "doing" {
			task.Contexts = append(task.Contexts[:i], task.Contexts[i+1:]...)
			break
		}
	}
	task.Completed = true
	task.CompletedDate = date
}

func ParseRecurrence(task *todotxt.Task) (time.Time, error) {
	isRepeatOnCompletion := false
	nextOpenTime := time.Time{}
	if task.HasAdditionalTags() {
		if v, ok := task.AdditionalTags[KeyRec]; ok {
			isRepeatOnCompletion = strings.HasSuffix(v, "*")
			if isRepeatOnCompletion {
				v = v[:len(v)-1]
			}
			num, err := strconv.Atoi(v[:len(v)-1])
			if err != nil {
				return nextOpenTime, err
			}
			period := v[len(v)-1:]
			if isRepeatOnCompletion && task.Completed {
				nextOpenTime = task.CompletedDate
			} else {
				nextOpenTime = task.CreatedDate
			}
			switch period {
			case "d":
				nextOpenTime = nextOpenTime.AddDate(0, 0, num)
			case "w":
				nextOpenTime = nextOpenTime.AddDate(0, 0, num*7)
			case "m":
				nextOpenTime = nextOpenTime.AddDate(0, num, 0)
			case "y":
				nextOpenTime = nextOpenTime.AddDate(num, 0, 0)
			default:
				return nextOpenTime, errors.New("invalid recurrence period")
			}
		}
	}
	return nextOpenTime, nil
}
