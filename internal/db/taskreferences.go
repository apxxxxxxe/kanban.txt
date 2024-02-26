package db

import (
	"github.com/1set/todotxt"
)

// TaskReferences is a list of tasks
// todotxt.TaskList is []todotxt.Task, but this is the []*todotxt.Task
type TaskReferences []*todotxt.Task

func NewTaskReferences() TaskReferences {
	return TaskReferences{}
}

func (tr *TaskReferences) AddTask(t *todotxt.Task) {
	*tr = append(*tr, t)
}

func (tr *TaskReferences) RemoveTask(t *todotxt.Task) {
	newList := NewTaskReferences()
	for i, task := range *tr {
		if task.String() != t.String() {
			newList.AddTask((*tr)[i])
		}
	}
	*tr = newList
}

func (tr *TaskReferences) Filter(pred todotxt.Predicate, preads ...todotxt.Predicate) *TaskReferences {
	combinedPred := []todotxt.Predicate{pred}
	combinedPred = append(combinedPred, preads...)

	result := NewTaskReferences()
	for i := range *tr {
		for _, pred := range combinedPred {
			if pred(*(*tr)[i]) {
				result.AddTask((*tr)[i])
				break
			}
		}
	}

	return &result
}

func (tr *TaskReferences) PredicateNot(pred todotxt.Predicate) todotxt.Predicate {
	return func(task todotxt.Task) bool {
		return !pred(task)
	}
}
