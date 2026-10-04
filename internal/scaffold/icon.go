package scaffold

import (
	"bytes"
	"hash/fnv"
	"image"
	"image/color"
	"image/png"
)

// PlaceholderIcon genera un PNG cuadrado con un degradado cuyo color depende de seed,
// para tener iconos válidos hasta reemplazarlos por los definitivos.
func PlaceholderIcon(size int, seed string) []byte {
	h := fnv.New32a()
	h.Write([]byte(seed))
	sum := h.Sum32()
	base := color.RGBA{uint8(40 + sum%140), uint8(40 + (sum>>8)%140), uint8(90 + (sum>>16)%140), 255}

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	radius := size / 6
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if outsideRounded(x, y, size, radius) {
				continue
			}
			f := float64(x+y) / float64(2*size)
			c := color.RGBA{shade(base.R, f), shade(base.G, f), shade(base.B, f), 255}
			// Marco interior más claro.
			inset := size / 4
			if x > inset && x < size-inset && y > inset && y < size-inset &&
				(x < inset+size/16 || x > size-inset-size/16 || y < inset+size/16 || y > size-inset-size/16) {
				c = color.RGBA{255, 255, 255, 220}
			}
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func shade(v uint8, f float64) uint8 {
	out := float64(v) * (1.15 - 0.4*f)
	if out > 255 {
		return 255
	}
	return uint8(out)
}

func outsideRounded(x, y, size, r int) bool {
	cx, cy := -1, -1
	switch {
	case x < r && y < r:
		cx, cy = r, r
	case x >= size-r && y < r:
		cx, cy = size-r-1, r
	case x < r && y >= size-r:
		cx, cy = r, size-r-1
	case x >= size-r && y >= size-r:
		cx, cy = size-r-1, size-r-1
	}
	if cx < 0 {
		return false
	}
	dx, dy := x-cx, y-cy
	return dx*dx+dy*dy > r*r
}
