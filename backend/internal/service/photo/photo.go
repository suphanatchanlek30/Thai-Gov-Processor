// Package photo implements service.PhotoService using bimg (libvips bindings).
package photo

import (
	"fmt"

	"github.com/h2non/bimg"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/preset"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/service"
)

// Range and step from docs/compression-algorithm.md: 66 candidate values, so
// binary search converges in at most 7 iterations (log2(66) ≈ 6.04).
const (
	minQuality = 30
	maxQuality = 95
)

type Service struct{}

func NewService() service.PhotoService {
	return Service{}
}

func (Service) ProcessPreset(input []byte, presetName string, custom *preset.Custom) (service.ProcessedImage, error) {
	width, height, maxKB, err := resolveTarget(presetName, custom)
	if err != nil {
		return service.ProcessedImage{}, err
	}

	// Crop to PNG first so every quality trial below re-encodes from a lossless
	// source instead of compounding JPEG artifacts from a prior lossy pass.
	cropped, err := bimg.NewImage(input).Process(bimg.Options{
		Width:         width,
		Height:        height,
		Crop:          true,
		Enlarge:       true,
		Gravity:       bimg.GravityCentre,
		Type:          bimg.PNG,
		StripMetadata: true,
	})
	if err != nil {
		return service.ProcessedImage{}, fmt.Errorf("photo: resize/crop: %w", err)
	}

	best, quality, found, err := compressToTarget(cropped, maxKB)
	if err != nil {
		return service.ProcessedImage{}, err
	}
	if !found {
		return service.ProcessedImage{}, fmt.Errorf("%w: %dx%d could not fit %d KB even at quality %d", service.ErrCannotMeetTarget, width, height, maxKB, minQuality)
	}

	size, err := bimg.NewImage(best).Size()
	if err != nil {
		return service.ProcessedImage{}, fmt.Errorf("photo: measure output: %w", err)
	}

	return service.ProcessedImage{
		Data:    best,
		Width:   size.Width,
		Height:  size.Height,
		SizeKB:  float64(len(best)) / 1024,
		Quality: quality,
	}, nil
}

func resolveTarget(presetName string, custom *preset.Custom) (width, height, maxKB int, err error) {
	if presetName == preset.CustomName {
		if custom == nil {
			return 0, 0, 0, fmt.Errorf("photo: custom preset requires width/height/max_kb")
		}
		return custom.Width, custom.Height, custom.MaxKB, nil
	}
	p, ok := preset.Lookup(presetName)
	if !ok {
		return 0, 0, 0, fmt.Errorf("photo: unknown preset %q", presetName)
	}
	return p.Width, p.Height, p.MaxKB, nil
}

// compressToTarget binary-searches JPEG quality for the highest value whose
// encoded size stays within maxKB, per docs/compression-algorithm.md.
func compressToTarget(pixelBuf []byte, maxKB int) (best []byte, bestQuality int, found bool, err error) {
	low, high := minQuality, maxQuality
	for low <= high {
		mid := (low + high) / 2
		out, encErr := bimg.NewImage(pixelBuf).Process(bimg.Options{
			Type:          bimg.JPEG,
			Quality:       mid,
			StripMetadata: true,
		})
		if encErr != nil {
			return nil, 0, false, fmt.Errorf("photo: encode at quality %d: %w", mid, encErr)
		}

		if float64(len(out))/1024 <= float64(maxKB) {
			best, bestQuality, found = out, mid, true
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return best, bestQuality, found, nil
}
