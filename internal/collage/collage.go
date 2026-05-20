// Copyright (c) 2026 Apache-2.0 license
//
// @file    collage.go
// @author  Karsh Ranjan
// @brief   Core grid-collage composition logic

package collage

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"path/filepath"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	xdraw "golang.org/x/image/draw"
)

// feat: cell label modes
//
// Four strategies for annotating each image in the grid. The chosen mode is
// passed in from the CLI via --labels. "grid" uses spreadsheet-style
// co-ordinates (1A, 1B …); "index" numbers images sequentially; "file" shows
// the source filename; "none" omits labels entirely.

// LabelMode controls how each cell is annotated.
type LabelMode int

const (
	LabelNone  LabelMode = iota
	LabelGrid            // "1A", "1B", "2A", …
	LabelIndex           // "1", "2", "3", …
	LabelFile            // bare filename
)

// ParseLabelMode converts a string flag value to a LabelMode.
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

// feat: automatic square-grid sizing
//
// Calculates the smallest rows × cols arrangement that holds all images whilst
// keeping the layout as square as possible. Achieved by rounding the square
// root of the image count up to the nearest integer for columns, then deriving
// rows accordingly. Spare cells at the end are left blank.

// SquareDimensions returns the smallest (rows, cols) pair whose product is ≥ n
// and whose shape is as close to square as possible.
func SquareDimensions(n int) (rows, cols int) {
	cols = int(math.Ceil(math.Sqrt(float64(n))))
	rows = int(math.Ceil(float64(n) / float64(cols)))
	return
}

// chore: visual palette and layout constants
//
// Fixed pixel measurements and RGBA colour values used throughout the
// composition. The dark-theme palette — near-black canvas, slightly lighter
// cell fills, and even darker label strips — keeps visual focus on the images
// themselves rather than the surrounding chrome.

// visual constants
const (
	cellPad   = 6  // gap between cells (and outer border)
	labelBarH = 28 // height of the label strip below each cell
)

var (
	colBg      = color.RGBA{0x1e, 0x1e, 0x1e, 0xff} // overall canvas background
	colCell    = color.RGBA{0x2d, 0x2d, 0x2d, 0xff} // empty cell background
	colLabelBg = color.RGBA{0x10, 0x10, 0x10, 0xff} // label strip background
	colLabelFg = color.RGBA{0xcc, 0xcc, 0xcc, 0xff} // label text colour
)

// Options holds optional rendering parameters for the collage.
type Options struct {
	BgColor  color.Color // canvas background; nil means use default dark theme
	FontSize float64     // label font size in points; 0 means default (16)
}

// ParseHexColor converts a hex color string (e.g. "#ff0000") to a color.Color.
// Returns nil for "transparent".
func ParseHexColor(s string) (color.Color, error) {
	if strings.ToLower(s) == "transparent" {
		return color.Transparent, nil
	}
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return nil, fmt.Errorf("invalid hex colour %q: expected 6-digit hex (e.g. ff0000)", s)
	}
	var r, g, b uint8
	_, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return nil, fmt.Errorf("invalid hex colour %q: %w", s, err)
	}
	return color.RGBA{r, g, b, 0xff}, nil
}

// feat: grid collage composition
//
// Assembles a single PNG canvas by placing each image into its grid slot.
// Cell dimensions are derived from the largest image in the set so that no
// image is ever cropped; smaller images are centred with letterboxing.
// An optional label bar is drawn beneath each cell. The indexOffset parameter
// lets multi-page runs carry sequential numbering across output files.

// Make builds a single grid collage image.
//
//   - imgs / filenames: parallel slices, len ≤ rows*cols (remaining cells are blank).
//   - indexOffset: added to the local index only for LabelIndex labels, so that
//     page 2 of a multi-collage run continues numbering from where page 1 left off.
func Make(imgs []image.Image, filenames []string, rows, cols int, mode LabelMode, indexOffset int, opts Options) *image.RGBA {
	if len(imgs) == 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	fontSize := opts.FontSize
	if fontSize <= 0 {
		fontSize = 16
	}
	barH := int(fontSize) + 12

	bg := opts.BgColor
	if bg == nil {
		bg = colBg
	}

	cellW, cellH := maxDims(imgs)

	rowH := cellH
	if mode != LabelNone {
		rowH += barH
	}

	totalW := cellPad + cols*(cellW+cellPad)
	totalH := cellPad + rows*(rowH+cellPad)

	dst := image.NewRGBA(image.Rect(0, 0, totalW, totalH))
	draw.Draw(dst, dst.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)

	for i, img := range imgs {
		r := i / cols
		c := i % cols

		x := cellPad + c*(cellW+cellPad)
		y := cellPad + r*(rowH+cellPad)

		// cell background (visible for letterbox areas)
		cellRect := image.Rect(x, y, x+cellW, y+cellH)
		draw.Draw(dst, cellRect, image.NewUniform(colCell), image.Point{}, draw.Src)

		// scale image to fill cell, preserving aspect ratio
		scaledInto(dst, img, cellRect)

		// optional label bar
		if mode != LabelNone {
			lbl := labelText(mode, i+indexOffset, r, c, filenames[i])
			renderLabel(dst, x, y+cellH, cellW, barH, fontSize, lbl)
		}
	}

	return dst
}

// feat: image scaling and letterboxing
//
// maxDims establishes a uniform cell size by taking the greatest width and
// height seen across the entire image set. scaledInto then draws each source
// image into that cell, shrinking or enlarging it whilst honouring its
// original aspect ratio — any leftover space shows the cell background colour.
// Bilinear interpolation (golang.org/x/image/draw) is used for quality.

// maxDims returns the maximum width and height across all images.
func maxDims(imgs []image.Image) (w, h int) {
	for _, img := range imgs {
		b := img.Bounds()
		if b.Dx() > w {
			w = b.Dx()
		}
		if b.Dy() > h {
			h = b.Dy()
		}
	}
	return
}

// scaledInto draws src centered and letterboxed inside dst's rect.
func scaledInto(dst *image.RGBA, src image.Image, rect image.Rectangle) {
	sb := src.Bounds()
	sw, sh := float64(sb.Dx()), float64(sb.Dy())
	dw, dh := float64(rect.Dx()), float64(rect.Dy())

	scale := math.Min(dw/sw, dh/sh)
	nw := int(sw * scale)
	nh := int(sh * scale)

	ox := rect.Min.X + (rect.Dx()-nw)/2
	oy := rect.Min.Y + (rect.Dy()-nh)/2

	xdraw.BiLinear.Scale(dst, image.Rect(ox, oy, ox+nw, oy+nh), src, sb, xdraw.Over, nil)
}

// feat: label text generation
//
// Produces the display string for a given cell depending on the active label
// mode. Grid-style labels pair a 1-based row number with a spreadsheet column
// letter (A–Z, then AA, AB …). The colLetters helper converts a zero-based
// column index using the bijective base-26 algorithm, matching the convention
// used in spreadsheet applications such as Excel.

// labelText returns the display string for one cell.
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

// colLetters converts a 0-based column index to spreadsheet-style letters
// (0→A, 25→Z, 26→AA, 27→AB, …).
func colLetters(col int) string {
	s := ""
	for col >= 0 {
		s = string(rune('A'+col%26)) + s
		col = col/26 - 1
	}
	return s
}

// feat: label bar rendering
//
// Draws a short strip of text directly below a cell onto the shared canvas.
// The strip background uses a darker shade than the cell itself for visual
// separation. Text is rendered with basicfont.Face7x13 (a 7-pixel-wide
// monospaced bitmap font from golang.org/x/image) and is truncated with an
// ellipsis if it would overflow the available width.

// renderLabel draws the label bar directly below a cell.
func renderLabel(dst *image.RGBA, x, y, width, barH int, fontSize float64, text string) {
	bar := image.Rect(x, y, x+width, y+barH)
	draw.Draw(dst, bar, image.NewUniform(colLabelBg), image.Point{}, draw.Src)

	face := loadFontFace(fontSize)
	if face == nil {
		return
	}
	defer face.Close()

	// estimate char width from font metrics
	adv, ok := face.GlyphAdvance('M')
	charW := 10
	if ok {
		charW = adv.Round()
	}
	if charW <= 0 {
		charW = int(fontSize * 0.6)
	}

	maxChars := (width - 8) / charW
	if maxChars <= 0 {
		return
	}
	runes := []rune(text)
	if len(runes) > maxChars {
		if maxChars > 3 {
			text = string(runes[:maxChars-3]) + "..."
		} else {
			text = string(runes[:maxChars])
		}
	}

	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(colLabelFg),
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.I(x + 4),
			Y: fixed.I(y + barH - 4),
		},
	}
	d.DrawString(text)
}

// loadFontFace returns an opentype font face at the given size.
func loadFontFace(size float64) font.Face {
	tt, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil
	}
	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil
	}
	return face
}
