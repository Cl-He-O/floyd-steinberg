package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strconv"

	"github.com/lucasb-eyer/go-colorful"
)

type RGBf struct {
	r, g, b float64
}

func dither(img image.Image, bpc int) *image.RGBA {
	errwindow := 2

	res := image.NewRGBA(img.Bounds())
	errbuf := [][]RGBf{}
	for i := 0; i < errwindow; i++ {
		errbuf = append(errbuf, make([]RGBf, img.Bounds().Dx()))
	}

	errdiff := func(x, y int) *RGBf {
		return &errbuf[y%errwindow][x]
	}

	nearest := func(x uint8) uint8 {
		k := float64(int(1<<bpc)-1) / 0xff
		return uint8(math.Round(float64(x)*k) / k)
	}

	errdiff_add := func(x int, y int, err_linear RGBf, k float64) {
		if x >= img.Bounds().Max.X || x < img.Bounds().Min.X {
			return
		}
		if y >= img.Bounds().Max.Y || y < img.Bounds().Min.Y {
			return
		}

		c := *errdiff(x, y)
		c.r += err_linear.r * k
		c.g += err_linear.g * k
		c.b += err_linear.b * k
		*errdiff(x, y) = c
	}

	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			var c_target colorful.Color
			var c_nearest colorful.Color
			{
				c, _ := colorful.MakeColor(img.At(x, y))
				r, g, b := c.LinearRgb()
				c_target = colorful.LinearRgb(r+errdiff(x, y).r, g+errdiff(x, y).g, b+errdiff(x, y).b)

				*errdiff(x, y) = RGBf{} // clear your buffer!!!

				R, G, B := c_target.Clamped().RGB255()

				cc := color.RGBA{nearest(R), nearest(G), nearest(B), 0xff} // i don't deal with alpha
				res.Set(x, y, cc)

				c_nearest, _ = colorful.MakeColor(cc)
			}

			{
				r, g, b := c_target.LinearRgb()
				rn, gn, bn := c_nearest.LinearRgb()

				err_linear := RGBf{r - rn, g - gn, b - bn}

				errdiff_add(x+1, y, err_linear, 7.0/16)
				errdiff_add(x+1, y+1, err_linear, 1.0/16)
				errdiff_add(x, y+1, err_linear, 5.0/16)
				errdiff_add(x-1, y+1, err_linear, 3.0/16)
			}
		}
	}

	return res
}

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	img, err := png.Decode(f)
	if err != nil {
		panic(err)
	}

	bpc, err := strconv.Atoi(os.Args[3])
	if err != nil {
		panic(err)
	}
	res := dither(img, bpc)

	f, err = os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	png.Encode(f, res)
}
