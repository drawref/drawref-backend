package drawref

import (
	"image"
	"image/color"
	"testing"

	"github.com/disintegration/imaging"
)

func TestThumbnailSettings(t *testing.T) {
	SetCachedThumbnailMinSizeKB(500)
	if GetThumbnailMinSizeKB() != 500 {
		t.Errorf("Expected 500, got %d", GetThumbnailMinSizeKB())
	}

	SetCachedThumbnailMinSizeKB(DefaultThumbnailMinFilesizeKB)
	if GetThumbnailMinSizeKB() != 400 {
		t.Errorf("Expected 400, got %d", GetThumbnailMinSizeKB())
	}
}

func TestAllowedThumbnailCacheSizes(t *testing.T) {
	if !IsAllowedThumbnailCacheSize(300) {
		t.Errorf("Expected 300 to be allowed in AllowedThumbnailCacheSizes")
	}
	if IsAllowedThumbnailCacheSize(287) {
		t.Errorf("Expected 287 not to be allowed in AllowedThumbnailCacheSizes")
	}
}

func TestImageFitResizing(t *testing.T) {
	// Create an 800x600 test image
	src := image.NewRGBA(image.Rect(0, 0, 800, 600))
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			src.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	// Fit with max=300
	maxVal := 300
	thumb := imaging.Fit(src, maxVal, maxVal, imaging.Lanczos)

	bounds := thumb.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w > maxVal || h > maxVal {
		t.Errorf("Thumbnail dimensions %dx%d exceed maxVal %d", w, h, maxVal)
	}

	// Original aspect ratio is 800:600 = 4:3 = 1.333
	// Expected thumb width is 300, height is 225
	if w != 300 {
		t.Errorf("Expected width 300, got %d", w)
	}
	if h != 225 {
		t.Errorf("Expected height 225, got %d", h)
	}
}

func TestSkipEncodingIfLargerThanOriginal(t *testing.T) {
	// Source image 200x150
	src := image.NewRGBA(image.Rect(0, 0, 200, 150))
	bounds := src.Bounds()
	origMax := max(bounds.Dx(), bounds.Dy())

	// Requested max=300 is larger than original max (200)
	reqMax := 300
	shouldSkip := reqMax >= origMax
	if !shouldSkip {
		t.Errorf("Expected shouldSkip to be true when reqMax %d >= origMax %d", reqMax, origMax)
	}

	// Requested max=150 is smaller than original max (200)
	reqMaxSmall := 150
	shouldSkipSmall := reqMaxSmall >= origMax
	if shouldSkipSmall {
		t.Errorf("Expected shouldSkip to be false when reqMax %d < origMax %d", reqMaxSmall, origMax)
	}
}
