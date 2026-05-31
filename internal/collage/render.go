package collage

import (
	"image"
	"math"

	xdraw "golang.org/x/image/draw"
)

// returns the maximum width and height across the image set
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

// draws src centred and letterboxed inside the destination rectangle
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
