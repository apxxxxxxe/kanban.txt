package tui

import (
	"time"

	"github.com/1set/todotxt"
)

const (
	todoProjects      = "Projects"
	todoPriority      = "Priority"
	todoTitle         = "Title"
	todoContexts      = "Contexts"
	todoDueDate       = "DueDate"
	todoCompletedDate = "CompletedDate"
	todoCreatedDate   = "CreatedDate"
)

func getTaskField(task *todotxt.Task, field string) string {
	switch field {
	case todoProjects:
		if len(task.Projects) == 0 {
			return ""
		}
		return task.Projects[0]
	case todoPriority:
		return task.Priority
	case todoTitle:
		return task.Todo
	case todoContexts:
		if len(task.Contexts) == 0 {
			return ""
		}
		return task.Contexts[0]
	case todoDueDate:
		return timeToStr(task.DueDate)
	case todoCompletedDate:
		return timeToStr(task.CompletedDate)
	case todoCreatedDate:
		return timeToStr(task.CreatedDate)
	default:
		panic("invalid field: " + field)
	}
}

func setTaskField(task *todotxt.Task, field, value string) {
	switch field {
	case todoProjects:
		if len(task.Projects) == 0 {
			task.Projects = []string{value}
		} else {
			task.Projects[0] = value
		}
	case todoPriority:
		task.Priority = value
	case todoTitle:
		task.Todo = value
	case todoContexts:
		if len(task.Contexts) == 0 {
			task.Contexts = []string{value}
		} else {
			task.Contexts[0] = value
		}
	case todoDueDate:
		task.DueDate = strToTime(value)
	case todoCompletedDate:
		task.CompletedDate = strToTime(value)
	case todoCreatedDate:
		task.CreatedDate = strToTime(value)
	default:
		panic("invalid field: " + field)
	}
}

func timeToStr(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(todotxt.DateLayout)
}

func strToTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(todotxt.DateLayout, s)
	if err != nil {
		panic(err)
	}
	return t
}
