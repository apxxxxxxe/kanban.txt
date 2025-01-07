package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/apxxxxxxe/kanban.txt/internal/db"
	"github.com/apxxxxxxe/kanban.txt/internal/tui"
)

func run() int {
	var (
		dataPath string
	)
	flag.StringVar(&dataPath, "data-path", "", "path of the data directory")
	flag.Parse()
	if dataPath != "" {
		db.CustomDataPath = dataPath
	}

	if err := tui.NewTui().Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
