package graph

import "image/color"

var (
	colors = []color.Color{
		color.NRGBA{0, 0, 128, 150},
		color.NRGBA{0, 128, 0, 150},
		color.NRGBA{128, 0, 0, 150},
		color.NRGBA{0, 128, 128, 150},
		color.NRGBA{128, 0, 128, 150},
		color.NRGBA{128, 128, 0, 150},
	}
)

func GetColor(pos int) color.Color {
	return colors[pos%len(colors)]
}
