package main

import (
	"fmt"
	"os"

	"github.com/apxxxxxxe/kanban.txt/internal/tui"
)

func run() int {
	if err := tui.NewTui().Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
