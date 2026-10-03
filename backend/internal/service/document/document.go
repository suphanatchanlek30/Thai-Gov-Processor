// Package document implements service.DocumentService using pdfcpu.
package document

import (
	"bytes"
	"fmt"
	"io"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/service"
)

const pdfContentType = "application/pdf"

type Service struct{}

func NewService() service.DocumentService {
	// Without this pdfcpu writes a config dir under $HOME and calls os.Exit(1)
	// when it can't, which kills the whole API on a read-only root filesystem.
	api.DisableConfigDir()
	return Service{}
}

// targetMaxKB stays informational: MergeRaw already runs pdfcpu's own
// OptimizeContext pass before writing, which is the only straightforward
// size reduction the library offers on a page-concatenation path.
func (Service) Merge(files []service.InputFile, targetMaxKB int) (service.MergedPDF, error) {
	if len(files) == 0 {
		return service.MergedPDF{}, fmt.Errorf("document: no input files")
	}

	conf := model.NewDefaultConfiguration()

	readers := make([]io.ReadSeeker, len(files))
	for i, f := range files {
		if f.ContentType == pdfContentType {
			readers[i] = bytes.NewReader(f.Data)
			continue
		}

		var page bytes.Buffer
		if err := api.ImportImages(nil, &page, []io.Reader{bytes.NewReader(f.Data)}, nil, conf); err != nil {
			return service.MergedPDF{}, fmt.Errorf("document: convert %q to PDF: %w", f.Name, err)
		}
		readers[i] = bytes.NewReader(page.Bytes())
	}

	var out bytes.Buffer
	if err := api.MergeRaw(readers, &out, false, conf); err != nil {
		return service.MergedPDF{}, fmt.Errorf("document: merge: %w", err)
	}

	pageCount, err := api.PageCount(bytes.NewReader(out.Bytes()), conf)
	if err != nil {
		return service.MergedPDF{}, fmt.Errorf("document: count pages: %w", err)
	}

	return service.MergedPDF{
		Data:       out.Bytes(),
		TotalPages: pageCount,
		SizeKB:     float64(out.Len()) / 1024,
	}, nil
}
