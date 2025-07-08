package images

import (
	"image"

	"github.com/disintegration/imaging"
)

func Fry(img image.Image) image.Image {
	fried := imaging.Blur(img, 0.8)
	fried = imaging.AdjustContrast(fried, 70)
	fried = imaging.AdjustBrightness(fried, 15)
	fried = imaging.AdjustSaturation(fried, 70)
	fried = imaging.Sharpen(fried, 30)
	fried = imaging.Resize(fried, fried.Bounds().Dx()/2, fried.Bounds().Dy()/2, imaging.NearestNeighbor)
	fried = imaging.Resize(fried, fried.Bounds().Dx()*2, fried.Bounds().Dy()*2, imaging.NearestNeighbor)

	fried = imaging.Convolve3x3(
		fried,
		[9]float64{
			-1, -1, 0,
			-1, 1, 1,
			0, 1, 1,
		},
		nil,
	)

	return fried
}
