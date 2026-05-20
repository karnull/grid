// Copyright (c) 2026 Apache-2.0 license
//
// @file    main.go
// @author  Karsh Ranjan
// @brief   Main entry point

package main

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"regexp"
	"text/template"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/karshPrime/grid/internal/collage"
	"github.com/spf13/cobra"
)

// feat: CLI definition and flags
//
// Declares the root cobra command and its optional flags. --rows and --cols let
// the user fix one or both grid dimensions; leaving both at zero triggers the
// automatic square layout. --labels selects one of the four annotation styles
// defined in the collage package. --background sets the canvas colour.
// --fontsize controls label text size.

var (
	flagRows       int
	flagCols       int
	flagLabels     string
	flagBackground string
	flagFontSize   float64
)

var rootCmd = &cobra.Command{
	Use:   "grid [images...]",
	Short: "Merge images into a grid collage",
	Long: `grid arranges a set of images into a PNG grid collage.

Default layout is square (e.g. 25 images → 5×5). Spare cells are left blank.
Output is saved as merged.png (or merged_1.png, merged_2.png, … when paging).

Examples:
  grid *.png
  grid --cols 3 --labels index --background transparent --fontsize 20 img1.png img2.png img3.png`,
	Args: cobra.MinimumNArgs(1),
	RunE: run,
}

func init() {
	cobra.AddTemplateFuncs(template.FuncMap{
		"stripTypes": stripTypes,
	})

	rootCmd.Flags().IntVar(&flagRows, "rows", 0, "number of rows (0 = auto)")
	rootCmd.Flags().IntVar(&flagCols, "cols", 0, "number of columns (0 = auto)")
	rootCmd.Flags().StringVar(&flagLabels, "labels", "none", "label style: none, grid, index, file")
	rootCmd.Flags().StringVar(&flagBackground, "background", "", "background colour: hex code (e.g. ff0000) or transparent")
	rootCmd.Flags().Float64Var(&flagFontSize, "fontsize", 16, "label font size in points")

	rootCmd.SetUsageTemplate(usageTemplate)
}

// stripTypes removes type annotations (int, string, float64, float) from pflag usage lines.
var reType = regexp.MustCompile(`(--\S+)\s+(int|string|float64|float)\s`)

func stripTypes(s string) string {
	return reType.ReplaceAllString(s, "$1   ")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// feat: command orchestration
//
// Validates flags, loads images, then resolves the final grid dimensions.
// When both --rows and --cols are given and their product is smaller than
// the total image count, the set is automatically split across numbered
// output files. Otherwise a single merged.png is produced.

// run is the cobra command handler.
func run(cmd *cobra.Command, args []string) error {
	mode, err := collage.ParseLabelMode(flagLabels)
	if err != nil {
		return err
	}

	opts := collage.Options{FontSize: flagFontSize}
	if flagBackground != "" {
		bg, err := collage.ParseHexColor(flagBackground)
		if err != nil {
			return err
		}
		opts.BgColor = bg
	}

	imgs, names := loadImages(args)
	if len(imgs) == 0 {
		return fmt.Errorf("no valid images found")
	}

	n := len(imgs)
	rows, cols := flagRows, flagCols
	bothSet := rows > 0 && cols > 0

	if bothSet && rows*cols < n {
		// Both dimensions fixed and too small: split into multiple pages.
		return writePages(imgs, names, rows, cols, mode, opts)
	}

	// Compute missing dimension(s).
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

	out := collage.Make(imgs, names, rows, cols, mode, 0, opts)
	if err := savePNG("merged.png", out); err != nil {
		return err
	}
	fmt.Println("Saved merged.png")
	return nil
}

// feat: multi-page collage output
//
// Partitions the full image list into fixed-size pages and writes each one
// to a separate file named merged_1.png, merged_2.png, and so on. The
// indexOffset passed to collage.Make ensures that sequential index labels
// (--labels index) continue numbering across all output files.

// writePages splits imgs across multiple collages of rows×cols each.
func writePages(imgs []image.Image, names []string, rows, cols int, mode collage.LabelMode, opts collage.Options) error {
	perPage := rows * cols
	numPages := int(math.Ceil(float64(len(imgs)) / float64(perPage)))

	for p := range numPages {
		start := p * perPage
		end := min(start+perPage, len(imgs))

		out := collage.Make(imgs[start:end], names[start:end], rows, cols, mode, start, opts)
		path := fmt.Sprintf("merged_%d.png", p+1)
		if err := savePNG(path, out); err != nil {
			return err
		}
		fmt.Printf("Saved %s (%d images)\n", path, end-start)
	}
	return nil
}

// feat: image loading with graceful degradation
//
// Opens and decodes each path supplied on the command line. Files that cannot
// be read or decoded emit a warning to stderr and are skipped rather than
// aborting the whole run. PNG, JPEG, and GIF are supported via the blank
// imports at the top; image.Decode selects the correct decoder automatically
// from the registered format table.

// loadImages decodes every path, warning and skipping on error.
func loadImages(paths []string) ([]image.Image, []string) {
	imgs := make([]image.Image, 0, len(paths))
	names := make([]string, 0, len(paths))

	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", path, err)
			continue
		}
		img, _, decErr := image.Decode(f)
		f.Close()
		if decErr != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", path, decErr)
			continue
		}
		imgs = append(imgs, img)
		names = append(names, path)
	}
	return imgs, names
}

// chore: PNG file output
//
// Creates (or truncates) the output file and passes the canvas to the standard
// library's PNG encoder. The file handle is closed via defer so the descriptor
// is always released, even when the encoder returns an error.

// savePNG encodes img to a new file at path.
func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// chore: custom usage template
//
// Removes type annotations (int, string, float64) from flag descriptions in
// the help output, giving a cleaner look.

var usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}{{$cmds := .Commands}}{{if eq (len .Groups) 0}}

Available Commands:{{range $cmds}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{else}}{{range $group := .Groups}}

{{.Title}}{{range $cmds}}{{if (and (eq .GroupID $group.ID) (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if not .AllChildCommandsHaveGroup}}

Additional Commands:{{range $cmds}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsagesWrapped 80 | trimTrailingWhitespaces | stripTypes}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsagesWrapped 80 | trimTrailingWhitespaces | stripTypes}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
