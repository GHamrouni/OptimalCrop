/*
Copyright (c) 2012, Ghassen Hamrouni <ghamrouni.iptech@gmail.com>

Permission to use, copy, modify, and/or distribute this software for any purpose
with or without fee is hereby granted, provided that the above copyright notice
and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF
THIS SOFTWARE.
*/

package optimalResize

import (
	"image"
	"imaging/resize"
	"math"
)

import _ "image/jpeg"

func OptimalResize(b image.Image, rectX int, rectY int, maxIter int) image.Image {

	bounds := b.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	aspect := float64(rectX) / float64(rectY)

	minInfoRatio := 0.85

	var W, H float64

	W = float64(width)
	H = float64(height)

	// Match the aspect ratio
	if aspect < 1.0 {
		H = aspect * W
	} else if aspect > 1 {
		W = (1.0 / aspect) * H
	} else {
		H = math.Min(H, W)
		W = H
	}

	// Crop the image to match the aspect ratio
	cropRegion := FindOptimalCropRegion(&b, int(H), int(W))

	// If initial cropping incur an information loss > 20%
	// crop the image to match the aspect ratio but don t
	// proceed to successive cropping
	if cropRegion.confidenceX < minInfoRatio || cropRegion.confidenceY < minInfoRatio {

		mpix := b.(*image.NRGBA)
		b = mpix.SubImage(cropRegion.region)
	}

	// Perform successive cropping if there is enough information left
	for maxIter > 0 && cropRegion.confidenceX > minInfoRatio &&
		cropRegion.confidenceY > minInfoRatio {

		mpix := b.(*image.NRGBA)
		b = mpix.SubImage(cropRegion.region)

		var w, h float64
		w, h = float64(b.Bounds().Max.X), float64(b.Bounds().Max.Y)

		cropRegion = FindOptimalCropRegion(&b, int(w*0.7), int(h*0.7))

		maxIter--
	}

	b = resize.Resize(uint(rectX), uint(rectY), b, resize.MitchellNetravali)
	return b
}
