package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/lucasb-eyer/go-colorful"
)

type RGBf struct {
	r, g, b float64
}

func main() {
	const bpc = 1 // bits per channel

	f, err := os.Open("gray.png")
	if err != nil {
		panic(err)
	}

	img, err := png.Decode(f)
	if err != nil {
		panic(err)
	}

	res := image.NewRGBA(img.Bounds())
	errdiff := [][]RGBf{make([]RGBf, img.Bounds().Dx()), make([]RGBf, img.Bounds().Dx())}

	nearest := func(x uint8) uint8 {
		k := 1.0 / 0xff * ((1 << bpc) - 1)
		return uint8(math.Round(float64(x)*k) / k)
	}

	errdiff_add := func(x int, y int, err_linear RGBf, k float64) {
		if x >= img.Bounds().Max.X || x < img.Bounds().Min.X {
			return
		}
		if y >= img.Bounds().Max.Y || y < img.Bounds().Min.Y {
			return
		}

		c := errdiff[y&0x1][x]
		c.r += err_linear.r * k
		c.g += err_linear.g * k
		c.b += err_linear.b * k
		errdiff[y&0x1][x] = c
	}

	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			var c_target colorful.Color
			var c_nearest colorful.Color
			{
				c, _ := colorful.MakeColor(img.At(x, y))
				r, g, b := c.LinearRgb() // adding error
				c_target = colorful.LinearRgb(r+errdiff[y&0x1][x].r, g+errdiff[y&0x1][x].g, b+errdiff[y&0x1][x].b)

				errdiff[y&0x1][x] = RGBf{} // clear your buffer!!!

				R, G, B := c_target.Clamped().RGB255()

				//if y > 60 && y < 100 && y%2 == 0 && x < 20 {
				//	fmt.Println(x, y, c_target, R, G, B)
				//}

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

	f, err = os.Create("out.png")
	if err != nil {
		panic(err)
	}
	png.Encode(f, res)
}
