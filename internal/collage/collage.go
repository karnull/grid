package collage

import (
	"image"
	"image/draw"
)

// composes a grid collage of the given images on a single canvas
func Make(imgs []image.Image, filenames []string, rows, cols int, mode LabelMode, indexOffset int, st *Style) *image.RGBA {
	if len(imgs) == 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	cellW, cellH := maxDims(imgs)

	rowH := cellH
	if mode != LabelNone {
		rowH += st.LabelBarH
	}

	totalW := st.Pad + cols*(cellW+st.Pad)
	totalH := st.Pad + rows*(rowH+st.Pad)

	dst := image.NewRGBA(image.Rect(0, 0, totalW, totalH))
	draw.Draw(dst, dst.Bounds(), image.NewUniform(st.Bg), image.Point{}, draw.Src)

	for i, img := range imgs {
		r := i / cols
		c := i % cols

		x := st.Pad + c*(cellW+st.Pad)
		y := st.Pad + r*(rowH+st.Pad)

		cellRect := image.Rect(x, y, x+cellW, y+cellH)
		draw.Draw(dst, cellRect, image.NewUniform(st.Cell), image.Point{}, draw.Src)
		scaledInto(dst, img, cellRect)

		if mode != LabelNone {
			lbl := labelText(mode, i+indexOffset, r, c, filenames[i])
			renderLabel(dst, st, x, y+cellH, cellW, lbl)
		}
	}

	return dst
}
