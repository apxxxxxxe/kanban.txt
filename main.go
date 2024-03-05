package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/1set/todotxt"
	"github.com/apxxxxxxe/kanban.txt/internal/tui"
)

func run() int {
	var (
		isCmdLine bool
	)
	flag.BoolVar(&isCmdLine, "cmd", false, "Run in command line mode")
	flag.Parse()

	if isCmdLine {
		if flag.NArg() == 0 {
			fmt.Fprintln(os.Stderr, "task string is required")
			return 1
		}
		task := todotxt.NewTask()
		task.Todo = strings.Join(flag.Args(), " ")
		if err := tui.NewTui().AddTask(task); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	} else {
		if err := tui.NewTui().Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	return 0
}

func main() {
	os.Exit(run())
}
