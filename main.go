package main

import (
	"image/color"
	"log"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/go-git/go-git/v5"
	"github.com/lusingander/gogigu"
)

const (
	appTitle = "fynegit"

	dateTimeFormat   = "2006/01/02 15:04:05"
	messageMaxLength = 80

	graphWidthUnit    = 30
	graphCircleRadius = 5
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
	w.Resize(fyne.NewSize(1200, 600))
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
	list := widget.NewList(
		func() int {
			return len(repo.Nodes)
		},
		func() fyne.CanvasObject {
			return commitGraphItem(repo)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			updateCommitGraphItem(repo, repo.Nodes[id], item)
		},
	)
	return list
}

func commitGraphItem(repo *gogigu.Repository) fyne.CanvasObject {
	graphAreaWidth := float32((getMaxPos(repo) + 1) * graphWidthUnit)
	graphArea := widget.NewLabel("")
	msg := widget.NewLabel("commit message")
	hash := widget.NewLabel("hash")
	author := widget.NewLabel("author")
	committedAt := widget.NewLabel("2006/01/02 15:04:05")
	var msgW, hashW, authorW float32 = 500, 100, 200
	graphArea.Move(fyne.NewPos(0, 0))
	msg.Move(fyne.NewPos(graphArea.Position().X+graphAreaWidth, 0))
	hash.Move(fyne.NewPos(msg.Position().X+msgW, 0))
	author.Move(fyne.NewPos(hash.Position().X+hashW, 0))
	committedAt.Move(fyne.NewPos(author.Position().X+authorW, 0))
	return container.NewWithoutLayout(
		graphArea,
		msg,
		hash,
		author,
		committedAt,
	)
}

func updateCommitGraphItem(repo *gogigu.Repository, node *gogigu.Node, item fyne.CanvasObject) {
	objs := item.(*fyne.Container).Objects
	objs[0] = calcCommitGraphTree(repo, node, item)
	objs[1].(*widget.Label).SetText(summaryMessage(node))
	objs[2].(*widget.Label).SetText(shortHash(node))
	objs[3].(*widget.Label).SetText(authorName(node))
	objs[4].(*widget.Label).SetText(commitedAt(node))
}

func calcCommitGraphTree(repo *gogigu.Repository, node *gogigu.Node, item fyne.CanvasObject) fyne.CanvasObject {
	graphAreaWidth := float32((getMaxPos(repo) + 1) * graphWidthUnit)
	graphAreaHeight := item.Size().Height
	rect := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	circle := &canvas.Circle{}
	circleRadius := float32(graphCircleRadius)
	circle.StrokeColor = color.NRGBA{0, 0, 128, 150}
	circle.FillColor = color.NRGBA{0, 0, 128, 50}
	circle.StrokeWidth = 2
	circlePosX := float32((node.Pos+1)*graphWidthUnit) - (graphWidthUnit / 2)
	circle.Move(fyne.NewPos(circlePosX-circleRadius/2, graphAreaHeight/2-circleRadius/2))
	circle.Resize(fyne.NewSize(circleRadius*2, circleRadius*2))
	graph := container.NewWithoutLayout(rect, circle)
	rect.Resize(fyne.NewSize(graphAreaWidth, graphAreaHeight))
	graph.Resize(fyne.NewSize(graphAreaWidth, graphAreaHeight))
	return graph
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

func authorName(node *gogigu.Node) string {
	return node.Commit.Author.Name
}

func commitedAt(node *gogigu.Node) string {
	return node.Commit.Committer.When.Format(dateTimeFormat)
}

func getMaxPos(repo *gogigu.Repository) int {
	max := 0
	for _, n := range repo.Nodes {
		if max < n.Pos {
			max = n.Pos
		}
	}
	return max
}
