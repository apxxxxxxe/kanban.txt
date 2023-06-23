package tui

import (
	"time"

	"github.com/1set/todotxt"
	"github.com/apxxxxxxe/kanban.txt/internal/task"
)

const (
	todoProjects      = "Projects"
	todoPriority      = "Priority"
	todoTitle         = "Title"
	todoContexts      = "Contexts"
	todoDueDate       = "DueDate"
	todoCompletedDate = "CompletedDate"
	todoCreatedDate   = "CreatedDate"
	todoRecurrence    = "Recurrence"
	todoNext          = "Next"
)

func getTaskField(t *todotxt.Task, field string) string {
	switch field {
	case todoProjects:
		if len(t.Projects) == 0 {
			return ""
		}
		return t.Projects[0]
	case todoPriority:
		return t.Priority
	case todoTitle:
		return t.Todo
	case todoContexts:
		if len(t.Contexts) == 0 {
			return ""
		}
		return t.Contexts[0]
	case todoDueDate:
		return timeToStr(t.DueDate)
	case todoCompletedDate:
		return timeToStr(t.CompletedDate)
	case todoCreatedDate:
		return timeToStr(t.CreatedDate)
	case todoRecurrence:
		return t.AdditionalTags[task.KeyRec]
	case todoNext:
		return t.AdditionalTags[task.KeyNext]
	default:
		panic("invalid field: " + field)
	}
}

func setTaskField(t *todotxt.Task, field, value string) {
	switch field {
	case todoProjects:
		if len(t.Projects) == 0 {
			t.Projects = []string{value}
		} else {
			t.Projects[0] = value
		}
	case todoPriority:
		t.Priority = value
	case todoTitle:
		t.Todo = value
	case todoContexts:
		if len(t.Contexts) == 0 {
			t.Contexts = []string{value}
		} else {
			t.Contexts[0] = value
		}
	case todoDueDate:
		t.DueDate = strToTime(value)
	case todoCompletedDate:
		t.CompletedDate = strToTime(value)
	case todoCreatedDate:
		t.CreatedDate = strToTime(value)
	case todoRecurrence:
		if t.AdditionalTags == nil {
			t.AdditionalTags = map[string]string{}
		}
		t.AdditionalTags[task.KeyRec] = value
	case todoNext:
		if t.AdditionalTags == nil {
			t.AdditionalTags = map[string]string{}
		}
		t.AdditionalTags[task.KeyNext] = value
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
