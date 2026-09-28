// Package stub provides temporary "not implemented" implementations of
// processor.PhotoProcessor and processor.PDFMerger.
//
// TODO(feat/image-pdf-processor): replace both constructors below with real
// implementations (resize/compress via a real image library, PDF merge via
// pdfcpu) and delete this package. main.go only needs its two wiring lines
// (NewNotImplementedPhotoProcessor / NewNotImplementedPDFMerger) changed to
// point at the new package — no handler code should need to change.
package stub

import (
	"fmt"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/preset"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/processor"
)

// ErrNotImplemented is returned by every method of the stub
// implementations. It wraps processor.ErrNotImplemented so handlers can
// match on that sentinel via errors.Is and map it to HTTP 501, regardless
// of which concrete processor implementation is wired in.
var ErrNotImplemented = fmt.Errorf("%w: processor not yet wired in (see feat/image-pdf-processor)", processor.ErrNotImplemented)

type notImplementedPhotoProcessor struct{}

// NewNotImplementedPhotoProcessor returns a processor.PhotoProcessor stub
// that always returns ErrNotImplemented. Swap this out in cmd/api/main.go
// for the real implementation once it lands.
func NewNotImplementedPhotoProcessor() processor.PhotoProcessor {
	return notImplementedPhotoProcessor{}
}

func (notImplementedPhotoProcessor) ProcessPreset(_ []byte, _ string, _ *preset.Custom) (processor.ProcessedImage, error) {
	return processor.ProcessedImage{}, ErrNotImplemented
}

type notImplementedPDFMerger struct{}

// NewNotImplementedPDFMerger returns a processor.PDFMerger stub that always
// returns ErrNotImplemented. Swap this out in cmd/api/main.go for the real
// implementation once it lands.
func NewNotImplementedPDFMerger() processor.PDFMerger {
	return notImplementedPDFMerger{}
}

func (notImplementedPDFMerger) Merge(_ []processor.InputFile, _ int) (processor.MergedPDF, error) {
	return processor.MergedPDF{}, ErrNotImplemented
}
