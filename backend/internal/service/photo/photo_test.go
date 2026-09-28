package photo

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"math/rand"
	"testing"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/preset"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/service"
)

// noisyJPEG produces high-entropy content that resists JPEG compression, so a
// resize to a small target still needs a real quality search to hit maxKB.
func noisyJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	r := rand.New(rand.NewSource(42))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{uint8(r.Intn(256)), uint8(r.Intn(256)), uint8(r.Intn(256)), 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	return buf.Bytes()
}

// solidJPEG produces flat, low-entropy content that compresses tiny at any quality.
func solidJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{200, 200, 200, 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	return buf.Bytes()
}

func TestProcessPreset_NeedsRealCompression(t *testing.T) {
	svc := NewService()
	input := noisyJPEG(t, 1200, 1400)

	result, err := svc.ProcessPreset(input, preset.CustomName, &preset.Custom{Width: 200, Height: 230, MaxKB: 10})
	if err != nil {
		t.Fatalf("ProcessPreset: %v", err)
	}

	if result.Width != 200 || result.Height != 230 {
		t.Errorf("dimensions = %dx%d, want 200x230", result.Width, result.Height)
	}
	if result.SizeKB > 10 {
		t.Errorf("SizeKB = %.2f, want <= 10", result.SizeKB)
	}
	if result.Quality < minQuality || result.Quality >= maxQuality {
		t.Errorf("Quality = %d, want a real compromise below the ceiling (%d..%d)", result.Quality, minQuality, maxQuality)
	}
	t.Logf("needs-compression case: quality=%d size=%.2fKB", result.Quality, result.SizeKB)
}

func TestProcessPreset_AlreadyUnderTarget(t *testing.T) {
	svc := NewService()
	input := solidJPEG(t, 50, 50)

	result, err := svc.ProcessPreset(input, preset.Passport, nil)
	if err != nil {
		t.Fatalf("ProcessPreset: %v", err)
	}

	passportPreset, _ := preset.Lookup(preset.Passport)
	if result.Width != passportPreset.Width || result.Height != passportPreset.Height {
		t.Errorf("dimensions = %dx%d, want %dx%d", result.Width, result.Height, passportPreset.Width, passportPreset.Height)
	}
	if result.Quality != maxQuality {
		t.Errorf("Quality = %d, want %d (max quality should already fit under target)", result.Quality, maxQuality)
	}
	if result.SizeKB >= float64(passportPreset.MaxKB) {
		t.Errorf("SizeKB = %.2f, want comfortably under %d", result.SizeKB, passportPreset.MaxKB)
	}
	t.Logf("already-under-target case: quality=%d size=%.2fKB (limit %dKB)", result.Quality, result.SizeKB, passportPreset.MaxKB)
}

func TestProcessPreset_CannotMeetTarget(t *testing.T) {
	svc := NewService()
	input := noisyJPEG(t, 1200, 1400)

	_, err := svc.ProcessPreset(input, preset.CustomName, &preset.Custom{Width: 200, Height: 230, MaxKB: 1})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, service.ErrCannotMeetTarget) {
		t.Fatalf("err = %v, want wrapping service.ErrCannotMeetTarget", err)
	}
}
