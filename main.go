package main

import (
	"log"
	"os"

	"fyne.io/fyne/v2/app"
	"github.com/lusingander/fynegit/internal/repository"
	"github.com/lusingander/fynegit/internal/ui"
)

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	repo, err := repository.OpenGitRepositoryFromArgs(args)
	if err != nil {
		return err
	}

	a := app.New()
	w := a.NewWindow("")
	ui.Start(w, repo)

	return nil
}
