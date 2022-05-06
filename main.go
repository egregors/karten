package main

import (
	"fmt"
	"os"

	"github.com/egregors/karten/cmd/add"
	"github.com/egregors/karten/cmd/learn"
	"github.com/egregors/karten/pkg/provider"
	"github.com/egregors/karten/pkg/store"
)

const cliAdd = "add"

func main() {
	// add mode
	if len(os.Args) > 1 && os.Args[1] == cliAdd {
		srv := add.NewSrv(
			store.CSV{Path: "words.csv"},
			provider.VerbFormen{URL: "https://www.verbformen.com/?w="},
			// todo: take debug config from CLI
			true,
		)

		if err := srv.Run(); err != nil {
			fmt.Println("ERR: ", err.Error())
			os.Exit(1)
		}

	}

	// learn mode
	srv, err := learn.NewSrv(
		store.CSV{Path: "words.csv"},
		true,
	)

	if err != nil {
		fmt.Println("ERR: ", err)
		os.Exit(1)
	}

	if err := srv.Run(); err != nil {
		fmt.Println("ERR: ", err)
		os.Exit(1)
	}
}
