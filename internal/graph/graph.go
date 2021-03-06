package graph

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/lusingander/fynegit/internal/gogigu"
	"github.com/lusingander/fynegit/internal/repository"
)

const (
	graphWidthUnit    = 20
	graphCircleRadius = 5
)

func CalcCommitGraphAreaWidth(rm *repository.RepositoryManager) float32 {
	return float32((rm.MaxPosX() + 1) * graphWidthUnit)
}

func CalcCommitGraphTreeRow(rm *repository.RepositoryManager, node *gogigu.Node, height float32) fyne.CanvasObject {
	graphAreaWidth := CalcCommitGraphAreaWidth(rm)
	graphAreaHeight := height

	posX := float32((node.PosX()+1)*graphWidthUnit) - (graphWidthUnit / 2)
	posY := float32(graphAreaHeight / 2)
	circleRadius := float32(graphCircleRadius)

	objs := make([]fyne.CanvasObject, 0)

	for _, edge := range rm.Edges(node.PosY()) {
		e := createCommitTreeEdge(node, edge, graphAreaWidth, graphAreaHeight, posX, posY, circleRadius)
		objs = append(objs, e)
	}

	rect := createDummyBackgroundRectangle(graphAreaWidth, graphAreaHeight)
	circle := createCommitObjectCircle(node, posX, posY, circleRadius)
	objs = append(objs, rect, circle)

	graph := container.NewWithoutLayout(objs...)
	graph.Resize(fyne.NewSize(graphAreaWidth, graphAreaHeight))

	return graph
}

func createCommitTreeEdge(node *gogigu.Node, edge *gogigu.Edge, graphAreaWidth, graphAreaHeight, posX, posY, circleRadius float32) *canvas.Line {
	switch edge.EdgeType {
	case gogigu.EdgeStraight:
		return createStraightEdge(
			edge,
			fyne.NewPos((float32(edge.PosX)+0.5)*graphWidthUnit, 0),
			graphAreaHeight,
		)
	case gogigu.EdgeUp:
		return createStraightEdge(
			edge,
			fyne.NewPos(posX, 0),
			posY-circleRadius,
		)
	case gogigu.EdgeDown:
		return createStraightEdge(
			edge,
			fyne.NewPos(posX, posY+circleRadius),
			posY-circleRadius,
		)
	case gogigu.EdgeBranch:
		return createEdge(
			edge,
			fyne.NewPos(posX+circleRadius, posY),
			float32((edge.PosX-node.PosX()-1)*graphWidthUnit+graphWidthUnit-int(circleRadius)),
			-graphAreaHeight/2,
		)
	case gogigu.EdgeMerge:
		return createEdge(
			edge,
			fyne.NewPos(posX+circleRadius, posY),
			float32((edge.PosX-node.PosX()-1)*graphWidthUnit+graphWidthUnit-int(circleRadius)),
			graphAreaHeight/2,
		)
	default:
		return nil
	}
}

func createEdge(edge *gogigu.Edge, leftOrTop fyne.Position, w, h float32) *canvas.Line {
	color := getColor(edge.PosX)
	e := &canvas.Line{
		StrokeColor: color,
		StrokeWidth: 2,
	}
	e.Move(leftOrTop)
	e.Resize(fyne.NewSize(w, h))
	return e
}

func createStraightEdge(edge *gogigu.Edge, leftOrTop fyne.Position, length float32) *canvas.Line {
	return createEdge(edge, leftOrTop, 0, length)
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
