// Package stub provides temporary "not implemented" implementations of
// service.PhotoService and service.DocumentService.
//
// TODO(feat/image-pdf-processor): replace both constructors below with real
// implementations (resize/compress via a real image library, PDF merge via
// pdfcpu) and delete this package. main.go only needs its two wiring lines
// (NewNotImplementedPhotoService / NewNotImplementedDocumentService) changed
// to point at the new package — no handler code should need to change.
package stub

import (
	"fmt"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/preset"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/service"
)

// ErrNotImplemented is returned by every method of the stub
// implementations. It wraps service.ErrNotImplemented so handlers can match
// on that sentinel via errors.Is and map it to HTTP 501, regardless of which
// concrete service implementation is wired in.
var ErrNotImplemented = fmt.Errorf("%w: service not yet wired in (see feat/image-pdf-processor)", service.ErrNotImplemented)

type notImplementedPhotoService struct{}

// NewNotImplementedPhotoService returns a service.PhotoService stub that
// always returns ErrNotImplemented. Swap this out in cmd/api/main.go for the
// real implementation once it lands.
func NewNotImplementedPhotoService() service.PhotoService {
	return notImplementedPhotoService{}
}

func (notImplementedPhotoService) ProcessPreset(_ []byte, _ string, _ *preset.Custom) (service.ProcessedImage, error) {
	return service.ProcessedImage{}, ErrNotImplemented
}

type notImplementedDocumentService struct{}

// NewNotImplementedDocumentService returns a service.DocumentService stub
// that always returns ErrNotImplemented. Swap this out in cmd/api/main.go
// for the real implementation once it lands.
func NewNotImplementedDocumentService() service.DocumentService {
	return notImplementedDocumentService{}
}

func (notImplementedDocumentService) Merge(_ []service.InputFile, _ int) (service.MergedPDF, error) {
	return service.MergedPDF{}, ErrNotImplemented
}
