package main

import (
	"flag"
	"fmt"
	"math"
	"os"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/karshPrime/grid/internal/collage"
	"github.com/karshPrime/grid/internal/config"
)

const usage = `Usage: grid [-r rows] [-c cols] [-l mode] [-o name] file...

Options:
  -r N    number of rows (0 = auto)
  -c N    number of columns (0 = auto)
  -l M    label mode: none, grid, index, file
  -o S    output file prefix (default: merged)
  -h      show this help

Configuration: ~/.config/grid.toml (or $XDG_CONFIG_HOME/grid.toml)
`

// program entry point
func main() {
	flag.CommandLine.SetOutput(os.Stderr)
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }

	rows := flag.Int("r", 0, "")
	cols := flag.Int("c", 0, "")
	labels := flag.String("l", "none", "")
	out := flag.String("o", "merged", "")
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(*rows, *cols, *labels, *out, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, "grid:", err)
		os.Exit(1)
	}
}

// resolves layout, loads images, and writes one or more collage files
func run(rows, cols int, labels, prefix string, paths []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	style, err := collage.StyleFromConfig(cfg)
	if err != nil {
		return err
	}
	mode, err := collage.ParseLabelMode(labels)
	if err != nil {
		return err
	}

	imgs, names := loadImages(paths)
	if len(imgs) == 0 {
		return fmt.Errorf("no valid images found")
	}

	n := len(imgs)
	bothSet := rows > 0 && cols > 0

	if bothSet && rows*cols < n {
		return writePages(imgs, names, rows, cols, mode, style, prefix)
	}

	if !bothSet {
		switch {
		case rows > 0:
			cols = int(math.Ceil(float64(n) / float64(rows)))
		case cols > 0:
			rows = int(math.Ceil(float64(n) / float64(cols)))
		default:
			rows, cols = collage.SquareDimensions(n)
		}
	}

	path := prefix + ".png"
	out := collage.Make(imgs, names, rows, cols, mode, 0, style)
	if err := savePNG(path, out); err != nil {
		return err
	}
	fmt.Printf("Saved %s\n", path)
	return nil
}
