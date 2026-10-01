package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strconv"
)

func main() {
	for _, size := range []int{192, 512} {
		f, err := os.Create("internal/web/static/icons/icon-" + strconv.Itoa(size) + ".png")
		if err != nil {
			panic(err)
		}
		if err := png.Encode(f, bunga(size)); err != nil {
			panic(err)
		}
		f.Close()
	}
}

// bunga menggambar ikon sakura: lima kelopak + pusat.
func bunga(size int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	cx, cy := float64(size)/2, float64(size)/2
	R := float64(size) * 0.21
	pr := float64(size) * 0.20
	cr := float64(size) * 0.11
	sakura := color.NRGBA{R: 0xEB, G: 0xA9, B: 0xBC, A: 255}
	paper := color.NRGBA{R: 0xFB, G: 0xF8, B: 0xF9, A: 255}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx, dy := float64(x)-cx, float64(y)-cy
			if math.Hypot(dx, dy) <= cr {
				img.SetNRGBA(x, y, paper)
				continue
			}
			for k := 0; k < 5; k++ {
				a := float64(k)*2*math.Pi/5 - math.Pi/2
				px, py := cx+R*math.Cos(a), cy+R*math.Sin(a)
				if math.Hypot(float64(x)-px, float64(y)-py) <= pr {
					img.SetNRGBA(x, y, sakura)
					break
				}
			}
		}
	}
	return img
}
