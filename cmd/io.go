package main

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"

	"github.com/karshPrime/grid/internal/collage"
)

// decodes images from paths, warning and skipping unreadable files
func loadImages(paths []string) ([]image.Image, []string) {
	imgs := make([]image.Image, 0, len(paths))
	names := make([]string, 0, len(paths))

	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grid: skipping %s: %v\n", path, err)
			continue
		}
		img, _, decErr := image.Decode(f)
		f.Close()
		if decErr != nil {
			fmt.Fprintf(os.Stderr, "grid: skipping %s: %v\n", path, decErr)
			continue
		}
		imgs = append(imgs, img)
		names = append(names, path)
	}
	return imgs, names
}

// splits the image set across multiple numbered pages
func writePages(imgs []image.Image, names []string, rows, cols int, mode collage.LabelMode, st *collage.Style, prefix string) error {
	perPage := rows * cols
	numPages := int(math.Ceil(float64(len(imgs)) / float64(perPage)))

	for p := 0; p < numPages; p++ {
		start := p * perPage
		end := start + perPage
		if end > len(imgs) {
			end = len(imgs)
		}

		out := collage.Make(imgs[start:end], names[start:end], rows, cols, mode, start, st)
		path := fmt.Sprintf("%s_%d.png", prefix, p+1)
		if err := savePNG(path, out); err != nil {
			return err
		}
		fmt.Printf("Saved %s (%d images)\n", path, end-start)
	}
	return nil
}

// encodes an image as PNG to the given path
func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
