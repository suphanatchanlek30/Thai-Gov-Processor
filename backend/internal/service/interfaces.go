// Package service defines the business-logic contracts between the HTTP
// handlers and the image/PDF libraries. PhotoService and DocumentService
// are implemented for real in feat/image-pdf-processor; until then
// internal/service/stub satisfies both interfaces.
package service

import (
	"errors"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/preset"
)

// Handlers match these via errors.Is regardless of which implementation is wired in.
var (
	ErrNotImplemented   = errors.New("service not implemented")
	ErrCannotMeetTarget = errors.New("cannot meet target size without unacceptable quality loss")
)

type ProcessedImage struct {
	Data    []byte
	Width   int
	Height  int
	SizeKB  float64
	Quality int
}

type InputFile struct {
	Name        string
	Data        []byte
	ContentType string
}

type MergedPDF struct {
	Data       []byte
	TotalPages int
	SizeKB     float64
}

// presetName is one of "ocsc", "passport", "teacher", "custom"; custom is
// required (and pre-validated) only for "custom".
type PhotoService interface {
	ProcessPreset(input []byte, presetName string, custom *preset.Custom) (ProcessedImage, error)
}

// Merge preserves file order; a multi-page PDF input contributes all its pages.
type DocumentService interface {
	Merge(files []InputFile, targetMaxKB int) (MergedPDF, error)
}
