// Package png allows for loading png images and applying
// image flitering effects on them.
package png

import(
	"image/color"
	// "fmt"
)

// Grayscale applies a grayscale filtering effect to the image
func (img *ImageTask) Grayscale(yStart int, yEnd int) {

	// Bounds returns defines the dimensions of the image. Always
	// use the bounds Min and Max fields to get Out the width
	// and height for the image
	bounds := img.Out.Bounds()
	for y := yStart; y < yEnd; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			//Returns the pixel (i.e., RGBA) value at a (x,y) position
			// Note: These get returned as int32 so based on the math you'll
			// be performing you'll need to do a conversion to float64(..)
			r, g, b, a := img.In.At(x, y).RGBA()

			//Note: The values for r,g,b,a for this assignment will range between [0, 65535].
			//For certain computations (i.e., convolution) the values might fall outside this
			// range so you need to clamp them between those values.
			greyC := clamp(float64(r+g+b) / 3)

			//Note: The values need to be stored back as uint16 (I know weird..but there's valid reasons
			// for this that I won't get into right now).
			img.Out.Set(x, y, color.RGBA64{greyC, greyC, greyC, uint16(a)})
		}
	}
}


func (img *ImageTask) Convolute(transform string, yStart int, yEnd int){

	// Attribution: http://www.songho.ca/dsp/convolution/convolution.html#convolution_2d

	kernel := make([][]float64, 3)

	// Set up kernel
	switch transform{
	case "S":
		kernel[0] = []float64{0,-1,0}
		kernel[1] = []float64{-1,5,-1}
		kernel[2] = []float64{0,-1,0}
	case "E":
		kernel[0] = []float64{-1,-1,-1}
		kernel[1] = []float64{-1,8,-1}
		kernel[2] = []float64{-1,-1,-1}
	case "B":
		kernel[0] = []float64{1 / 9.0, 1 / 9, 1 / 9.0}
		kernel[1] = []float64{1 / 9.0, 1 / 9.0, 1 / 9.0}
		kernel[2] = []float64{1 / 9.0, 1 / 9.0, 1 / 9.0}
	case "G":
		img.Grayscale(yStart, yEnd)
		return
	}

	kRows := len(kernel)
	kCols := len(kernel[0])

	bounds := img.Out.Bounds()
	for x := bounds.Min.X; x < bounds.Max.X; x++{
		for y := yStart; y < yEnd; y++{
			r, g, b, a := img.In.At(x, y).RGBA()
			var rC, gC, bC float64

			for m := 0; m < kRows; m ++{
				for n:= 0; n < kCols; n++{

					// get indexes for adjacent pixels
					ii := x + m - 1
					jj := y + n - 1

					r, g, b, a = img.In.At(ii, jj).RGBA()

					//Zero Padding
					if ii >= bounds.Min.X && ii < bounds.Max.X && jj >= bounds.Min.Y && jj < bounds.Max.Y {
						rC += float64(r) * kernel[m][n]
						gC += float64(g) * kernel[m][n]
						bC += float64(b) * kernel[m][n]
					}
				}
			}
			
			newRGB := color.RGBA64{clamp(rC), clamp(gC), clamp(bC), uint16(a)}
			img.Out.Set(x, y, newRGB)
		}
	}

}