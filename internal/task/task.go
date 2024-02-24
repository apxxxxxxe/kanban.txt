package task

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/1set/todotxt"
)

const (
	KeyRec        = "rec" // 繰り返し情報
	KeyNote       = "note" // 備考
	KeyStartDoing = "doing" // Doingにした日時
)

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

func ToDone(task *todotxt.Task) {
	for i, c := range task.Contexts {
		if c == "doing" {
			task.Contexts = append(task.Contexts[:i], task.Contexts[i+1:]...)
			break
		}
	}
	task.Complete()
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
			var dur time.Duration
			switch period {
			case "d":
				dur = time.Duration(num) * 24 * time.Hour
			case "w":
				dur = time.Duration(num) * 7 * 24 * time.Hour
			case "m":
				dur = time.Duration(num) * 30 * 24 * time.Hour
			case "y":
				dur = time.Duration(num) * 365 * 24 * time.Hour
			default:
				return nextOpenTime, errors.New("invalid recurrence period")
			}
			if isRepeatOnCompletion && task.Completed {
				nextOpenTime = task.CompletedDate.Add(dur)
			} else {
				nextOpenTime = task.CreatedDate.Add(dur)
			}
		}
	}
	return nextOpenTime, nil
}
