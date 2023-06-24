package task

import (
	"errors"
	"strconv"
	"time"

	"github.com/1set/todotxt"
)

const (
	KeyRec  = "rec"
)

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

func ParseRecurrence(task *todotxt.Task) (time.Time, error) {
	nextOpenTime := time.Time{}
	if task.HasAdditionalTags() {
		if v, ok := task.AdditionalTags[KeyRec]; ok {
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
			nextOpenTime = task.CreatedDate.Add(dur)
		}
	}
	return nextOpenTime, nil
}
