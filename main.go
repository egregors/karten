package main

import (
	"fmt"
	"os"

	"github.com/egregors/karten/cmd"
)

const (
	cliAdd = "add"
)

func main() {
	// todo: get words file path from the CLI args or ENV
	p := "words.csv"

	// todo: create empty CSV if not exist

	// add mode
	if len(os.Args) > 1 && os.Args[1] == cliAdd {
		if err := cmd.RunAddModeCLI(p); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	// learn mode
	if err := cmd.RunCLI(p); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
