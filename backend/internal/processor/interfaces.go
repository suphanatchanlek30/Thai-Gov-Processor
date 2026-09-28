// Package processor defines the contracts for photo resizing/compression
// and PDF merging. This file is the ground truth for the follow-up
// "feat/image-pdf-processor" PR: it implements PhotoProcessor and PDFMerger
// for real (resize + binary-search JPEG quality search, pdfcpu-based
// merge). Handlers in internal/handler only depend on these interfaces, so
// swapping the stub package in internal/processor/stub for a real
// implementation is a one-line change in cmd/api/main.go.
package processor

import (
	"errors"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/preset"
)

// Sentinel errors that handlers (internal/handler) match on with errors.Is
// to pick the right HTTP status, so both the stub and the real
// implementation can share the same handler code path.
var (
	// ErrNotImplemented means the processor is still the temporary stub
	// (internal/processor/stub). Handlers map this to HTTP 501.
	ErrNotImplemented = errors.New("processor not implemented")

	// ErrCannotMeetTarget means no JPEG quality in the allowed range (or no
	// merge strategy) got the output under the requested max_kb without an
	// unacceptable quality loss. Handlers map this to HTTP 422.
	ErrCannotMeetTarget = errors.New("cannot meet target size without unacceptable quality loss")
)

// ProcessedImage is the result of running a photo through a preset:
// resized/cropped to the preset's dimensions and JPEG-compressed to fit
// under the preset's max_kb via binary search over quality 30-95 (see
// docs/compression-algorithm.md).
type ProcessedImage struct {
	Data    []byte  // final JPEG bytes
	Width   int     // actual output width in px
	Height  int     // actual output height in px
	SizeKB  float64 // final size in KB
	Quality int     // JPEG quality that was selected (30-95)
}

// InputFile is one uploaded file (image or PDF) to be merged into a single
// output PDF, in the order it should appear.
type InputFile struct {
	Name        string // original filename, for error messages/logging
	Data        []byte
	ContentType string // as declared by the multipart upload
}

// MergedPDF is the result of combining one or more InputFile values into a
// single PDF, preserving page order across files (a multi-page PDF input
// contributes all of its pages, not just the first).
type MergedPDF struct {
	Data       []byte
	TotalPages int
	SizeKB     float64
}

// PhotoProcessor resizes/crops an input image to a preset's target
// dimensions and compresses it (binary search on JPEG quality) so the
// output does not exceed the preset's max_kb.
//
// input is the raw uploaded image bytes (JPEG/PNG/WEBP). presetName is one
// of "ocsc", "passport", "teacher", "custom". When presetName is "custom",
// custom must be non-nil and already validated (preset.ValidateCustom);
// for the three built-in presets, custom is ignored and may be nil.
//
// Implementations should return an error wrapping a 422-appropriate
// condition (via the caller mapping it through internal/apierr) when no
// quality in the allowed range fits under max_kb.
type PhotoProcessor interface {
	ProcessPreset(input []byte, presetName string, custom *preset.Custom) (ProcessedImage, error)
}

// PDFMerger combines images and/or existing PDFs into a single PDF, in the
// order given by files, attempting to keep the result under targetMaxKB.
type PDFMerger interface {
	Merge(files []InputFile, targetMaxKB int) (MergedPDF, error)
}
