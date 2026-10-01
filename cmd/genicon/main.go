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
		if err := png.Encode(f, sakura(size)); err != nil {
			panic(err)
		}
		f.Close()
	}
}

// sakura menggambar ikon: tile hijau gelap + bunga sakura 5 kelopak runcing + pusat.
func sakura(size int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	tile := color.NRGBA{R: 0x16, G: 0x30, B: 0x2A, A: 255}
	kelopak := color.NRGBA{R: 0xEB, G: 0xA9, B: 0xBC, A: 255}
	putih := color.NRGBA{R: 0xFB, G: 0xF8, B: 0xF9, A: 255}
	stamen := color.NRGBA{R: 0xA9, G: 0x41, B: 0x6A, A: 255}

	cx, cy := float64(size)/2, float64(size)/2
	rtip := float64(size) * 0.36
	rinner := float64(size) * 0.10
	wbase := float64(size) * 0.14
	rpusat := float64(size) * 0.085
	rstamen := float64(size) * 0.035

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// tile penuh (maskable)
			img.SetNRGBA(x, y, tile)

			dx, dy := float64(x)-cx, float64(y)-cy
			r := math.Hypot(dx, dy)
			if r <= rpusat {
				img.SetNRGBA(x, y, putih)
				if r <= rstamen {
					img.SetNRGBA(x, y, stamen)
				}
				continue
			}
			if r > rtip {
				continue
			}
			phi := math.Atan2(dy, dx)
			w := wbase * (rtip - r) / (rtip - rinner)
			halfAng := w / r
			for k := 0; k < 5; k++ {
				theta := -math.Pi/2 + float64(k)*2*math.Pi/5
				d := math.Abs(phi - theta)
				if d > math.Pi {
					d = 2*math.Pi - d
				}
				if d <= halfAng {
					img.SetNRGBA(x, y, kelopak)
					break
				}
			}
		}
	}
	return img
}
