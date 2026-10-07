package alike

import (
	"bytes"
	"image"
	"math/bits"

	"github.com/disintegration/imaging"

	// Register the decoders the source CDNs actually serve.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

// dHash sample grid: a 9x8 grayscale image yields 8 horizontal comparisons per
// row × 8 rows = 64 bits.
const (
	hashWidth  = 9
	hashHeight = 8
)

// DecodePosterHash decodes an image and returns its difference hash. The hash is
// robust to re-encoding, mild cropping and watermarks, so the same scene's
// framegrab thumbnail hashes similarly across different sites.
func DecodePosterHash(data []byte) (uint64, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	return dHash(img), nil
}

// dHash computes a 64-bit difference hash: each bit records whether a pixel is
// brighter than its right-hand neighbour in a downscaled grayscale image.
func dHash(img image.Image) uint64 {
	// imaging.Grayscale returns an NRGBA where R=G=B, so the R channel is the
	// grayscale value.
	small := imaging.Grayscale(imaging.Resize(img, hashWidth, hashHeight, imaging.Lanczos))

	var hash uint64
	bit := 0
	for y := 0; y < hashHeight; y++ {
		for x := 0; x < hashWidth-1; x++ {
			left := small.NRGBAAt(x, y).R
			right := small.NRGBAAt(x+1, y).R
			if left > right {
				hash |= 1 << uint(bit)
			}
			bit++
		}
	}
	return hash
}

// HammingSimilarity converts the bit distance between two hashes into a 0..1
// similarity (1 = identical). A hash of 0 means "not computed"; comparisons
// involving it return 0 so a missing poster contributes no signal.
func HammingSimilarity(a, b uint64) float64 {
	if a == 0 || b == 0 {
		return 0
	}
	distance := bits.OnesCount64(a ^ b)
	return 1 - float64(distance)/64
}
