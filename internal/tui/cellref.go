package tui

import todo "github.com/1set/todotxt"

type FeedCellRef struct {
	Feed   *todo.Task
	Cursor int
}

func NewFeedCellRef(f *todo.Task) *FeedCellRef {
	return &FeedCellRef{
		Feed:   f,
		Cursor: 0,
	}
}
