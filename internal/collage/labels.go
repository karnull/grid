package collage

import (
	"fmt"
	"image"
	"image/draw"
	"path/filepath"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// LabelMode controls how each cell is annotated.
type LabelMode int

const (
	LabelNone LabelMode = iota
	LabelGrid
	LabelIndex
	LabelFile
)

// converts a string flag value to a LabelMode
func ParseLabelMode(s string) (LabelMode, error) {
	switch strings.ToLower(s) {
	case "none", "":
		return LabelNone, nil
	case "grid":
		return LabelGrid, nil
	case "index":
		return LabelIndex, nil
	case "file":
		return LabelFile, nil
	default:
		return LabelNone, fmt.Errorf("invalid label mode %q: choose none, grid, index, or file", s)
	}
}

// returns the display string for a single cell
func labelText(mode LabelMode, globalIdx, row, col int, filename string) string {
	switch mode {
	case LabelGrid:
		return fmt.Sprintf("%d%s", row+1, colLetters(col))
	case LabelIndex:
		return fmt.Sprintf("%d", globalIdx+1)
	case LabelFile:
		return filepath.Base(filename)
	default:
		return ""
	}
}

// converts a 0-based column index to spreadsheet-style letters
func colLetters(col int) string {
	s := ""
	for col >= 0 {
		s = string(rune('A'+col%26)) + s
		col = col/26 - 1
	}
	return s
}

// draws a label strip below a cell
func renderLabel(dst *image.RGBA, st *Style, x, y, width int, text string) {
	bar := image.Rect(x, y, x+width, y+st.LabelBarH)
	draw.Draw(dst, bar, image.NewUniform(st.LabelBg), image.Point{}, draw.Src)

	if st.Face == nil {
		return
	}

	avail := fixed.I(width - 8)
	if avail <= 0 {
		return
	}
	text = truncate(st.Face, text, avail)

	w := font.MeasureString(st.Face, text)
	metrics := st.Face.Metrics()
	tx := x + (width-w.Round())/2
	ty := y + (st.LabelBarH+metrics.Ascent.Round()-metrics.Descent.Round())/2

	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(st.LabelFg),
		Face: st.Face,
		Dot:  fixed.Point26_6{X: fixed.I(tx), Y: fixed.I(ty)},
	}
	d.DrawString(text)
}

// shortens text with an ellipsis until it fits within max
func truncate(face font.Face, text string, max fixed.Int26_6) string {
	if font.MeasureString(face, text) <= max {
		return text
	}
	runes := []rune(text)
	for len(runes) > 0 {
		candidate := string(runes) + "..."
		if font.MeasureString(face, candidate) <= max {
			return candidate
		}
		runes = runes[:len(runes)-1]
	}
	return ""
}
