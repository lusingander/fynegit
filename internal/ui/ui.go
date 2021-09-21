package ui

import (
	"image/color"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/lusingander/fynegit/internal/graph"
	"github.com/lusingander/fynegit/internal/repository"
	"github.com/lusingander/gogigu"
)

const (
	dateTimeFormat   = "2006/01/02 15:04:05"
	messageMaxLength = 80

	graphWidthUnit    = 30
	graphCircleRadius = 5
)

var (
	defaultWindowSize = fyne.NewSize(1200, 600)
)

type manager struct {
	fyne.Window
	repo *gogigu.Repository
}

func Start(w fyne.Window, repo *gogigu.Repository) {
	m := &manager{w, repo}
	if repo == nil {
		m.SetContent(m.buildEmptyView())
	} else {
		m.SetContent(m.buildCommitGraphView())
	}
	m.Resize(defaultWindowSize)
	m.ShowAndRun()
}

func (m *manager) buildEmptyView() fyne.CanvasObject {
	open := func() {
		callback := func(lu fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, m.Window)
				return
			}
			if lu == nil {
				return // canceled
			}
			repo, err := repository.OpenGitRepository(lu.String()[7:]) // `file://`
			if err != nil {
				dialog.ShowError(err, m.Window)
				return
			}
			m.repo = repo
			m.SetContent(m.buildCommitGraphView())
		}
		dialog.ShowFolderOpen(callback, m.Window)
	}
	openButton := widget.NewButtonWithIcon("Open Git Repository", theme.StorageIcon(), open)
	return container.NewCenter(openButton)
}

func (m *manager) buildCommitGraphView() fyne.CanvasObject {
	if m.repo == nil {
		log.Fatalln("m.repo must not be nil")
	}
	list := widget.NewList(
		func() int {
			return len(m.repo.Nodes)
		},
		func() fyne.CanvasObject {
			return commitGraphItem(m.repo)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			updateCommitGraphItem(m.repo, m.repo.Nodes[id], item)
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

func updateCommitGraphItem(repo *gogigu.Repository, node *gogigu.Node, item fyne.CanvasObject) {
	objs := item.(*fyne.Container).Objects
	objs[0] = calcCommitGraphTree(repo, node, item)
	objs[1].(*widget.Label).SetText(summaryMessage(node))
	objs[2].(*widget.Label).SetText(shortHash(node))
	objs[3].(*widget.Label).SetText(authorName(node))
	objs[4].(*widget.Label).SetText(commitedAt(node))
}

func calcCommitGraphTree(repo *gogigu.Repository, node *gogigu.Node, item fyne.CanvasObject) fyne.CanvasObject {
	graphAreaWidth := float32((repo.MaxPosX() + 1) * graphWidthUnit)
	graphAreaHeight := item.Size().Height

	posX := float32((node.PosX()+1)*graphWidthUnit) - (graphWidthUnit / 2)
	posY := float32(graphAreaHeight / 2)
	circleRadius := float32(graphCircleRadius)

	objs := make([]fyne.CanvasObject, 0)
	for _, edge := range repo.Edges(node.PosY()) {
		createEdge := func(leftOrTop fyne.Position, length float32, vertical bool) *canvas.Line {
			e := canvas.NewLine(graph.GetColor(edge.PosX))
			e.StrokeWidth = 2
			e.Move(leftOrTop)
			if vertical {
				e.Resize(fyne.NewSize(0, length))
			} else {
				e.Resize(fyne.NewSize(length, 0))
			}
			return e
		}
		switch edge.EdgeType {
		case gogigu.EdgeStraight:
			e := createEdge(
				fyne.NewPos((float32(edge.PosX)+0.5)*graphWidthUnit, 0),
				graphAreaHeight,
				true,
			)
			objs = append(objs, e)
		case gogigu.EdgeUp:
			e := createEdge(
				fyne.NewPos(posX, 0),
				posY-circleRadius,
				true,
			)
			objs = append(objs, e)
		case gogigu.EdgeDown:
			e := createEdge(
				fyne.NewPos(posX, posY+circleRadius),
				posY-circleRadius,
				true,
			)
			objs = append(objs, e)
		case gogigu.EdgeBranch:
			e1 := createEdge(
				fyne.NewPos(posX+circleRadius, posY),
				float32((edge.PosX-node.PosX()-1)*graphWidthUnit+graphWidthUnit-int(circleRadius)),
				false,
			)
			e2 := createEdge(
				fyne.NewPos((float32(edge.PosX)+0.5)*graphWidthUnit, 0),
				graphAreaHeight/2,
				true,
			)
			objs = append(objs, e1, e2)
		case gogigu.EdgeMerge:
			e1 := createEdge(
				fyne.NewPos(posX+circleRadius, posY),
				float32((edge.PosX-node.PosX()-1)*graphWidthUnit+graphWidthUnit-int(circleRadius)),
				false,
			)
			e2 := createEdge(
				fyne.NewPos((float32(edge.PosX)+0.5)*graphWidthUnit, graphAreaHeight/2),
				graphAreaHeight/2,
				true,
			)
			objs = append(objs, e1, e2)
		}
	}

	rect := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	rect.Resize(fyne.NewSize(graphAreaWidth, graphAreaHeight))

	circle := &canvas.Circle{}
	circle.StrokeColor = graph.GetColor(node.PosX())
	circle.FillColor = graph.GetColor(node.PosX())
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
	return node.Commit.Author.When.Format(dateTimeFormat)
}
