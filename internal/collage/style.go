package collage

import (
	"fmt"
	"image/color"
	"os"

	"github.com/karshPrime/grid/internal/config"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

// Style holds resolved rendering parameters used by Make.
type Style struct {
	Bg        color.Color
	Cell      color.Color
	LabelBg   color.Color
	LabelFg   color.Color
	Pad       int
	LabelBarH int
	FontSize  float64
	Face      font.Face
}

// builds a Style from a loaded configuration
func StyleFromConfig(c *config.Config) (*Style, error) {
	bg, err := config.ParseColor(c.Background)
	if err != nil {
		return nil, fmt.Errorf("background: %w", err)
	}
	cell, err := config.ParseColor(c.Cell)
	if err != nil {
		return nil, fmt.Errorf("cell: %w", err)
	}
	lbg, err := config.ParseColor(c.LabelBg)
	if err != nil {
		return nil, fmt.Errorf("label_bg: %w", err)
	}
	lfg, err := config.ParseColor(c.LabelFg)
	if err != nil {
		return nil, fmt.Errorf("label_fg: %w", err)
	}

	size := c.FontSize
	if size <= 0 {
		size = 13
	}
	face, err := loadFace(c.Font, size)
	if err != nil {
		return nil, err
	}

	bar := c.LabelBarHeight
	if bar <= 0 {
		bar = int(size) + 8
	}
	pad := c.CellPad
	if pad < 0 {
		pad = 0
	}

	return &Style{
		Bg:        bg,
		Cell:      cell,
		LabelBg:   lbg,
		LabelFg:   lfg,
		Pad:       pad,
		LabelBarH: bar,
		FontSize:  size,
		Face:      face,
	}, nil
}

// loads a TTF/OTF font face, falling back to the embedded Go Regular face
func loadFace(path string, size float64) (font.Face, error) {
	var data []byte
	if path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("font: %w", err)
		}
		data = b
	} else {
		data = goregular.TTF
	}
	parsed, err := opentype.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("font parse: %w", err)
	}
	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		Size: size,
		DPI:  72,
	})
	if err != nil {
		return nil, fmt.Errorf("font face: %w", err)
	}
	return face, nil
}
