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
	"math"
)

import _ "image/jpeg"

// Represents a region to be cropped in the input image
type croppingArea struct {
	region image.Rectangle
	/// Represents the preserved information (on x-axis) When
	/// the image is cropped
	confidenceX float64
	/// Represents the preserved information (on y-axis) When
	// the image is cropped
	confidenceY float64
}

// Find the max subinterval with length = intervalSize
// It takes O(n) operations to find the interval
func FindMaxSubInterval(data []float64, intervalSize int) (T int, preservedInfo float64) {

	n := len(data)
	cumulativeData := make([]float64, n)

	cumulativeData[0] = data[0]

	// Compute the cumulative information
	// this data structure is used to accelerate
	// computations
	for i := 1; i < n; i++ {
		cumulativeData[i] = cumulativeData[i-1] + data[i]
	}

	T = 0

	var maxSum int64
	var depth int

	maxSum = 0
	depth = 0

	// Find an optimal interval in O(n)
	for i := intervalSize; i < n; i++ {
		var sum int64
		sum = 0

		if (i - intervalSize) == 0 {
			sum = int64(cumulativeData[i])
		} else {
			sum = int64(cumulativeData[i] - cumulativeData[i-intervalSize-1])
		}

		if sum > maxSum {
			maxSum = sum
			T = i - intervalSize
			depth = 0
		} else if sum == maxSum {
			depth++
		} else {
			if depth > 0 {
				T = T + depth/2
				depth = 0
			}
		}
	}

	if depth > 0 {
		T = T + depth/2
	}

	preservedInfo = float64(maxSum) / cumulativeData[n-1]

	return
}

func CalulatePixelIntensity(m *image.Image, x int, y int) int {
	r, g, b, _ := (*m).At(x, y).RGBA()

	// I is the intensity of the pixel
	// measured with Manhattan distance
	return int(255.0 * (r + g + b) / (65535.0 * 3.0))
}

func FindOptimalCropRegion(m *image.Image, rectX int, rectY int) croppingArea {

	bounds := (*m).Bounds()

	width, height := bounds.Max.X, bounds.Max.Y

	// The histogram of the image
	var H [256]float64

	// The self-informations of lines/columns
	Hx := make([]float64, width)
	Hy := make([]float64, height)

	for y := 0; y < 256; y++ {
		H[y] = 0
	}

	for i := 0; i < width; i++ {
		Hx[i] = 0
	}

	for i := 0; i < height; i++ {
		Hy[i] = 0
	}

	// Compute the histogram of the image
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {

			// I is the intensity of the pixel
			// measured with Manhattan distance
			I := CalulatePixelIntensity(m, x, y)

			// Increment the histogram
			H[I] += 1
		}
	}

	//
	// Normalize the histogram
	//
	sum := 0.0

	for y := 0; y < 256; y++ {
		sum += H[y]
	}

	for y := 0; y < 256; y++ {
		H[y] = H[y] / sum
	}

	// Compute the self-information for line/columns
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {

			I := CalulatePixelIntensity(m, x, y)

			// H[I] = the probability of the pixel
			// -log(p) = the information in the pixel
			Hx[x] += -math.Log(H[I])
			Hy[y] += -math.Log(H[I])
		}
	}

	// The x-coordinate of the optimal
	// cropping region
	Tx, preservedInfoX := FindMaxSubInterval(Hx, rectX)

	// The y-coordinate of the optimal
	// cropping region
	Ty, preservedInfoY := FindMaxSubInterval(Hy, rectY)

	if rectX >= width {
		preservedInfoX = 1
	}

	if rectY >= height {
		preservedInfoY = 1
	}

	// Return the solution
	return croppingArea{
		region:      image.Rect(Tx, Ty, Tx + rectY, Ty + rectX),
		confidenceX: preservedInfoX,
		confidenceY: preservedInfoY}
}
