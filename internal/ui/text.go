package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func textSize(s string) fyne.Size {
	return fyne.MeasureText(s, theme.TextSize(), fyne.TextStyle{})
}

func textWidth(s string) float32 {
	return textSize(s).Width
}

func cutText(s string, maxWidth, lowerBuf, upperBuf float32, try int) string {
	rs := []rune(s)
	lower := 0
	upper := len(rs)
	ptr := upper / 2
	c := 0
	for {
		s := string(rs[0:ptr])
		w := textWidth(s)
		if maxWidth-lowerBuf <= w && w <= maxWidth+upperBuf {
			return s
		} else if w < maxWidth-lowerBuf {
			lower = ptr
			ptr = (ptr + upper) / 2
		} else { // if maxWidth < w+upperBuf
			upper = ptr
			ptr = (lower + ptr) / 2
		}
		c += 1
		if c >= try {
			return s
		}
	}
}

func dummyPaddingSpaces(w float32) string {
	if w <= 0 {
		return ""
	}
	buf := textWidth(" ")
	pad := strings.Repeat(" ", 120)
	return cutText(pad, w, buf, buf, 8)
}
