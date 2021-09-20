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
	edges := calcEdges(repo)
	list := widget.NewList(
		func() int {
			return len(repo.Nodes)
		},
		func() fyne.CanvasObject {
			return commitGraphItem(repo)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			updateCommitGraphItem(edges, repo, repo.Nodes[id], item)
		},
	)
	return list
}

func commitGraphItem(repo *gogigu.Repository) fyne.CanvasObject {
	graphAreaWidth := float32((repo.MaxPosX() + 1) * graphWidthUnit)
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

func updateCommitGraphItem(edges map[int][]*edge, repo *gogigu.Repository, node *gogigu.Node, item fyne.CanvasObject) {
	objs := item.(*fyne.Container).Objects
	objs[0] = calcCommitGraphTree(edges, repo, node, item)
	objs[1].(*widget.Label).SetText(summaryMessage(node))
	objs[2].(*widget.Label).SetText(shortHash(node))
	objs[3].(*widget.Label).SetText(authorName(node))
	objs[4].(*widget.Label).SetText(commitedAt(node))
}

func calcCommitGraphTree(edges map[int][]*edge, repo *gogigu.Repository, node *gogigu.Node, item fyne.CanvasObject) fyne.CanvasObject {
	graphAreaWidth := float32((repo.MaxPosX() + 1) * graphWidthUnit)
	graphAreaHeight := item.Size().Height

	posX := float32((node.PosX()+1)*graphWidthUnit) - (graphWidthUnit / 2)
	posY := float32(graphAreaHeight / 2)
	circleRadius := float32(graphCircleRadius)

	objs := make([]fyne.CanvasObject, 0)
	for _, edge := range edges[node.PosY()] {
		switch edge.edgeType {
		case straight:
			e := canvas.NewLine(color.NRGBA{0, 0, 128, 150})
			e.StrokeWidth = 2
			e.Move(fyne.NewPos((float32(edge.posX)+0.5)*graphWidthUnit, 0))
			e.Resize(fyne.NewSize(0, graphAreaHeight))
			objs = append(objs, e)
		case up:
			e := canvas.NewLine(color.NRGBA{0, 0, 128, 150})
			e.StrokeWidth = 2
			e.Move(fyne.NewPos(posX, 0))
			e.Resize(fyne.NewSize(0, posY-circleRadius))
			objs = append(objs, e)
		case down:
			e := canvas.NewLine(color.NRGBA{0, 0, 128, 150})
			e.StrokeWidth = 2
			e.Move(fyne.NewPos(posX, posY+circleRadius))
			e.Resize(fyne.NewSize(0, posY-circleRadius))
			objs = append(objs, e)
		case branch:
			e1 := canvas.NewLine(color.NRGBA{0, 0, 128, 150})
			e1.StrokeWidth = 2
			e1.Move(fyne.NewPos(posX+circleRadius, posY))
			e1.Resize(fyne.NewSize(float32((edge.posX-node.PosX()-1)*graphWidthUnit+graphWidthUnit-int(circleRadius)), 0))
			e2 := canvas.NewLine(color.NRGBA{0, 0, 128, 150})
			e2.StrokeWidth = 2
			e2.Move(fyne.NewPos((float32(edge.posX)+0.5)*graphWidthUnit, 0))
			e2.Resize(fyne.NewSize(0, graphAreaHeight/2))
			objs = append(objs, e1, e2)
		case merge:
			e1 := canvas.NewLine(color.NRGBA{0, 0, 128, 150})
			e1.StrokeWidth = 2
			e1.Move(fyne.NewPos(posX+circleRadius, posY))
			e1.Resize(fyne.NewSize(float32((edge.posX-node.PosX()-1)*graphWidthUnit+graphWidthUnit-int(circleRadius)), 0))
			e2 := canvas.NewLine(color.NRGBA{0, 0, 128, 150})
			e2.StrokeWidth = 2
			e2.Move(fyne.NewPos((float32(edge.posX)+0.5)*graphWidthUnit, graphAreaHeight/2))
			e2.Resize(fyne.NewSize(0, graphAreaHeight/2))
			objs = append(objs, e1, e2)
		}
	}

	rect := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	rect.Resize(fyne.NewSize(graphAreaWidth, graphAreaHeight))

	circle := &canvas.Circle{}
	circle.StrokeColor = color.NRGBA{0, 0, 128, 150}
	circle.FillColor = color.NRGBA{0, 0, 128, 50}
	circle.StrokeWidth = 2
	circle.Move(fyne.NewPos(posX-circleRadius, posY-circleRadius))
	circle.Resize(fyne.NewSize(circleRadius*2, circleRadius*2))

	objs = append(objs, rect, circle)

	graph := container.NewWithoutLayout(objs...)
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
	return node.ShortHash()
}

func authorName(node *gogigu.Node) string {
	return node.Commit.Author.Name
}

func commitedAt(node *gogigu.Node) string {
	return node.Commit.Committer.When.Format(dateTimeFormat)
}

type edgeType int

const (
	straight edgeType = iota
	up
	down
	branch
	merge
)

type edge struct {
	edgeType
	posX int
}

func calcEdges(repo *gogigu.Repository) map[int][]*edge {
	edges := make(map[int][]*edge)
	for i := range repo.Nodes {
		edges[i] = make([]*edge, 0)
	}
	for _, n := range repo.Nodes {
		h := n.Hash()
		for _, child := range repo.Children(h) {
			edges[n.PosY()] = append(edges[n.PosY()], &edge{up, n.PosX()})
			if n.PosX() == child.PosX() {
				for y := n.PosY() - 1; y > child.PosY(); y-- {
					edges[y] = append(edges[y], &edge{straight, n.PosX()})
				}
			} else if n.PosX() < child.PosX() {
				edges[n.PosY()] = append(edges[n.PosY()], &edge{branch, child.PosX()})
				for y := n.PosY() - 1; y > child.PosY(); y-- {
					edges[y] = append(edges[y], &edge{straight, child.PosX()})
				}
			}
		}
		for _, parent := range repo.Parents(h) {
			edges[n.PosY()] = append(edges[n.PosY()], &edge{down, n.PosX()})
			if n.PosX() < parent.PosX() {
				edges[n.PosY()] = append(edges[n.PosY()], &edge{merge, parent.PosX()})
				for y := n.PosY() + 1; y < parent.PosY(); y++ {
					edges[y] = append(edges[y], &edge{straight, parent.PosX()})
				}
			}
		}
	}
	return edges
}
