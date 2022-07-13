package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/egregors/karten/cmd/add"
	"github.com/egregors/karten/cmd/learn"
	"github.com/egregors/karten/pkg/provider"
	"github.com/egregors/karten/pkg/store"
	"github.com/jessevdk/go-flags"
)

// Server is runnable service
type Server interface {
	Run() error
}

// Opts is App settings (from cli args or ENV)
type Opts struct {
	Add bool `short:"a" long:"add" description:"Run add-mode to add new word in your collection"`
	Dbg bool `long:"dbg" env:"DEBUG" description:"Debug mode"`
}

func main() {
	var opts Opts
	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("cli error: %v", err)
		}
		os.Exit(2)
	}

	storage, err := makeStorage()
	if err != nil {
		fmt.Printf("can't make a storage: %s", err.Error())
	}

	var srv Server

	if opts.Add { // run add-mode
		srv = add.NewSrv(
			storage,
			provider.VerbFormen{URL: "https://www.verbformen.com/?w="},
			opts.Dbg,
		)
	} else { // learn mode
		srv, err = learn.NewSrv(
			storage,
			opts.Dbg,
		)

		if err != nil {
			fmt.Printf("can't make server: %s\n", err)
			os.Exit(1)
		}

	}

	if err := srv.Run(); err != nil {
		fmt.Println("ERR: ", err)
		os.Exit(1)
	}
}

func makeStorage() (*store.CSV, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("can't get home dir: %w", err)
	}
	path := filepath.Join(home, ".karten", "words.csv")

	if err != nil {
		return nil, fmt.Errorf("can't get CSV path: %w", err)
	}

	storage, err := store.NewCSV(path)
	if err != nil {
		return nil, fmt.Errorf("can't create store: %w", err)
	}

	return storage, nil
}
