package main

import (
	"log"
	"os"

	"fyne.io/fyne/v2/app"
	"github.com/lusingander/fynegit/internal/repository"
	"github.com/lusingander/fynegit/internal/ui"
	"github.com/lusingander/gogigu"
)

const (
	appTitle = "fynegit"
)

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	repo, err := openGitRepositoryFromArgs(args)
	if err != nil {
		return err
	}

	a := app.New()
	w := a.NewWindow(appTitle)
	ui.Start(w, repo)

	return nil
}

func openGitRepositoryFromArgs(args []string) (*gogigu.Repository, error) {
	if len(args) <= 1 {
		return nil, nil
	}
	path := args[1]
	repo, err := repository.OpenGitRepository(path)
	if err != nil {
		return nil, err
	}
	return repo, nil
}
