package main

import (
	"log"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/go-git/go-git/v5"
	"github.com/lusingander/gogigu"
)

const (
	appTitle = "fynegit"

	dateTimeFormat   = "2006/01/02 15:04:05"
	messageMaxLength = 80
)

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	path := args[1] // fixme
	repo, err := openGitRepository(path)
	if err != nil {
		return err
	}

	a := app.New()
	w := a.NewWindow(appTitle)
	w.SetContent(buildCommitGraphView(repo))
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()

	return nil
}

func openGitRepository(path string) (*gogigu.Repository, error) {
	src, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	repo, err := gogigu.Calculate(src)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func buildCommitGraphView(repo *gogigu.Repository) fyne.CanvasObject {
	return widget.NewList(
		func() int {
			return len(repo.Nodes)
		},
		func() fyne.CanvasObject {
			return commitGraphItem()
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			updateCommitGraphItem(repo.Nodes[id], item)
		},
	)
}

func commitGraphItem() fyne.CanvasObject {
	msg := widget.NewLabel("commit message")
	hash := widget.NewLabel("hash")
	committedAt := widget.NewLabel("2006/01/02 15:04:05")
	return container.NewHBox(
		msg,
		hash,
		committedAt,
	)
}

func updateCommitGraphItem(node *gogigu.Node, item fyne.CanvasObject) {
	objs := item.(*fyne.Container).Objects
	objs[0].(*widget.Label).SetText(summaryMessage(node))
	objs[1].(*widget.Label).SetText(shortHash(node))
	objs[2].(*widget.Label).SetText(commitedAt(node))
}

func summaryMessage(node *gogigu.Node) string {
	msg := strings.Split(node.Commit.Message, "\n")[0]
	if len(msg) > messageMaxLength {
		msg = msg[:messageMaxLength]
	}
	return msg
}

func shortHash(node *gogigu.Node) string {
	return node.Hash[:7]
}

func commitedAt(node *gogigu.Node) string {
	return node.Commit.Committer.When.Format(dateTimeFormat)
}
