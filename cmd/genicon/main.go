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

// sakura menggambar bunga sakura lima kelopak meruncing + pusat.
func sakura(size int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	cx, cy := float64(size)/2, float64(size)/2
	rtip := float64(size) * 0.44   // ujung kelopak
	rinner := float64(size) * 0.10 // pangkal kelopak / jari-jari pusat
	wbase := float64(size) * 0.14  // setengah-lebar kelopak di pangkal

	kelopak := color.NRGBA{R: 0xEB, G: 0xA9, B: 0xBC, A: 255}
	pusat := color.NRGBA{R: 0xFB, G: 0xF8, B: 0xF9, A: 255}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx, dy := float64(x)-cx, float64(y)-cy
			r := math.Hypot(dx, dy)
			if r <= rinner {
				img.SetNRGBA(x, y, pusat)
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
