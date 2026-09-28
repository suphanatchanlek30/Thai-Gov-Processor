// Package stub provides "not implemented" placeholders for
// service.PhotoService and service.DocumentService until
// feat/image-pdf-processor lands. cmd/api/main.go's two wiring lines are
// the only thing that need to change to swap them out.
package stub

import (
	"fmt"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/preset"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/service"
)

var ErrNotImplemented = fmt.Errorf("%w: service not yet wired in (see feat/image-pdf-processor)", service.ErrNotImplemented)

type notImplementedPhotoService struct{}

func NewNotImplementedPhotoService() service.PhotoService {
	return notImplementedPhotoService{}
}

func (notImplementedPhotoService) ProcessPreset(_ []byte, _ string, _ *preset.Custom) (service.ProcessedImage, error) {
	return service.ProcessedImage{}, ErrNotImplemented
}

type notImplementedDocumentService struct{}

func NewNotImplementedDocumentService() service.DocumentService {
	return notImplementedDocumentService{}
}

func (notImplementedDocumentService) Merge(_ []service.InputFile, _ int) (service.MergedPDF, error) {
	return service.MergedPDF{}, ErrNotImplemented
}
