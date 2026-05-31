package config

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds user-tunable visual settings loaded from grid.toml.
type Config struct {
	Background     string  `toml:"background"`
	Cell           string  `toml:"cell"`
	LabelBg        string  `toml:"label_bg"`
	LabelFg        string  `toml:"label_fg"`
	Font           string  `toml:"font"`
	FontSize       float64 `toml:"font_size"`
	CellPad        int     `toml:"cell_pad"`
	LabelBarHeight int     `toml:"label_bar_height"`
}

// returns the default visual settings
func Defaults() *Config {
	return &Config{
		Background:     "#1e1e1e",
		Cell:           "#2d2d2d",
		LabelBg:        "#101010",
		LabelFg:        "#cccccc",
		Font:           "",
		FontSize:       13,
		CellPad:        6,
		LabelBarHeight: 0,
	}
}

// resolves the default config path under the user's home directory
func DefaultPath() string {
	if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, "grid.toml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "grid.toml")
}

// loads the global config, falling back to defaults when the file is absent
func Load() (*Config, error) {
	cfg := Defaults()
	path := DefaultPath()
	if path == "" {
		return cfg, nil
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	} else if err != nil {
		return nil, err
	}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return cfg, nil
}

// parses a hex colour like #rrggbb, #rrggbbaa, or the keyword "transparent"
func ParseColor(s string) (color.Color, error) {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "transparent") {
		return color.Transparent, nil
	}
	s = strings.TrimPrefix(s, "#")
	switch len(s) {
	case 6:
		var r, g, b uint8
		if _, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b); err != nil {
			return nil, fmt.Errorf("invalid hex colour %q", s)
		}
		return color.RGBA{r, g, b, 0xff}, nil
	case 8:
		var r, g, b, a uint8
		if _, err := fmt.Sscanf(s, "%02x%02x%02x%02x", &r, &g, &b, &a); err != nil {
			return nil, fmt.Errorf("invalid hex colour %q", s)
		}
		return color.RGBA{r, g, b, a}, nil
	}
	return nil, fmt.Errorf("invalid hex colour %q: expected 6 or 8 hex digits", s)
}
