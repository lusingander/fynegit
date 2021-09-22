package graph

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/lusingander/gogigu"
)

const (
	graphWidthUnit    = 30
	graphCircleRadius = 5
)

func CalcCommitGraphAreaWidth(repo *gogigu.Repository) float32 {
	return float32((repo.MaxPosX() + 1) * graphWidthUnit)
}

func CalcCommitGraphTreeRow(repo *gogigu.Repository, node *gogigu.Node, height float32) fyne.CanvasObject {
	graphAreaWidth := CalcCommitGraphAreaWidth(repo)
	graphAreaHeight := height

	posX := float32((node.PosX()+1)*graphWidthUnit) - (graphWidthUnit / 2)
	posY := float32(graphAreaHeight / 2)
	circleRadius := float32(graphCircleRadius)

	objs := make([]fyne.CanvasObject, 0)
	for _, edge := range repo.Edges(node.PosY()) {
		createEdge := func(leftOrTop fyne.Position, length float32, vertical bool) *canvas.Line {
			e := canvas.NewLine(getColor(edge.PosX))
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

	rect := createDummyBackgroundRectangle(graphAreaWidth, graphAreaHeight)
	circle := createCommitObjectCircle(node, posX, posY, circleRadius)
	objs = append(objs, rect, circle)

	graph := container.NewWithoutLayout(objs...)
	graph.Resize(fyne.NewSize(graphAreaWidth, graphAreaHeight))

	return graph
}

func createDummyBackgroundRectangle(width, height float32) fyne.CanvasObject {
	color := color.Transparent
	rect := &canvas.Rectangle{
		StrokeColor: color,
		FillColor:   color,
	}
	rect.Resize(fyne.NewSize(width, height))
	return rect
}

func createCommitObjectCircle(node *gogigu.Node, posX, posY, circleRadius float32) fyne.CanvasObject {
	color := getColor(node.PosX())
	circle := &canvas.Circle{
		StrokeColor: color,
		FillColor:   color,
		StrokeWidth: 2,
	}
	circle.Move(fyne.NewPos(posX-circleRadius, posY-circleRadius))
	circle.Resize(fyne.NewSize(circleRadius*2, circleRadius*2))
	return circle
}
