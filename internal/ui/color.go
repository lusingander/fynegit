package ui

import (
	"image/color"

	"github.com/lusingander/fynegit/internal/repository"
)

var (
	refsTagColorBg    = color.NRGBA{255, 255, 100, 150}
	refsTagColorFg    = color.NRGBA{100, 100, 0, 255}
	refsBranchColorBg = color.NRGBA{150, 200, 150, 150}
	refsBranchColorFg = color.NRGBA{0, 90, 0, 255}
	refsRemoteColorBg = color.NRGBA{200, 150, 200, 150}
	refsRemoteColorFg = color.NRGBA{70, 0, 70, 255}

	refsNoticeColorBg = color.NRGBA{200, 200, 200, 200}
	refsNoticeColorFg = color.NRGBA{100, 100, 100, 255}
)

func refsColor(t repository.RefType) (color.Color, color.Color) {
	switch t {
	case repository.Tag:
		return refsTagColorBg, refsTagColorFg
	case repository.Branch:
		return refsBranchColorBg, refsBranchColorFg
	case repository.RemoteBranch:
		return refsRemoteColorBg, refsRemoteColorFg
	}
	return nil, nil
}

func refsNoticeColor() (color.Color, color.Color) {
	return refsNoticeColorBg, refsNoticeColorFg
}
