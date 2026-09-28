package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/apierr"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/processor"
)

const defaultTargetMaxKB = 500

// MergePDF handles POST /api/v1/documents/merge-pdf: validates each
// uploaded file (image or PDF), runs them through the PDFMerger in the
// order they were submitted, stores the result, and returns a presigned
// download URL.
//
// While PDFMerger is the stub implementation this always returns 501 after
// passing validation — validation itself is fully enforced today.
func (h *Handler) MergePDF(c *gin.Context) {
	maxBytes := h.MaxUploadMB * 1024 * 1024
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)

	targetMaxKB := defaultTargetMaxKB
	if raw := c.PostForm("target_max_kb"); raw != "" {
		v, ok := parsePositiveIntForm(raw)
		if !ok {
			apierr.BadRequest(c, "target_max_kb must be a positive integer")
			return
		}
		targetMaxKB = v
	}

	form, err := c.MultipartForm()
	if err != nil {
		if isMaxBytesError(err) {
			apierr.PayloadTooLarge(c, fmt.Sprintf("upload exceeds maximum size of %d MB", h.MaxUploadMB))
			return
		}
		apierr.BadRequest(c, "expected multipart/form-data with files[]")
		return
	}

	fileHeaders := form.File["files[]"]
	if len(fileHeaders) == 0 {
		apierr.BadRequest(c, "files[] is required: provide one or more image/PDF files")
		return
	}

	inputs := make([]processor.InputFile, 0, len(fileHeaders))
	for _, fh := range fileHeaders {
		f, err := fh.Open()
		if err != nil {
			apierr.BadRequest(c, fmt.Sprintf("could not read uploaded file %q", fh.Filename))
			return
		}
		data, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			if isMaxBytesError(err) {
				apierr.PayloadTooLarge(c, fmt.Sprintf("upload exceeds maximum size of %d MB", h.MaxUploadMB))
				return
			}
			apierr.BadRequest(c, fmt.Sprintf("could not read uploaded file %q", fh.Filename))
			return
		}

		contentType := detectContentType(data)
		if !isAllowedMergeInputType(contentType) {
			apierr.BadRequest(c, fmt.Sprintf("unsupported file type %q for %q: only JPG, PNG, WEBP and PDF are supported (HEIC is not supported)", contentType, fh.Filename))
			return
		}

		inputs = append(inputs, processor.InputFile{
			Name:        fh.Filename,
			Data:        data,
			ContentType: contentType,
		})
	}

	result, err := h.PDFs.Merge(inputs, targetMaxKB)
	if err != nil {
		writeProcessorError(c, err)
		return
	}

	filename := fmt.Sprintf("merged_%s.pdf", randomID())
	key := "processed/" + filename

	ctx := c.Request.Context()
	if err := h.Storage.Put(ctx, key, result.Data, pdfContentType); err != nil {
		apierr.Internal(c, "failed to store merged file")
		return
	}
	downloadURL, err := h.Storage.PresignGet(ctx, key, h.PresignTTL)
	if err != nil {
		apierr.Internal(c, "failed to generate download URL")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"filename":     filename,
		"total_pages":  result.TotalPages,
		"size_kb":      result.SizeKB,
		"download_url": downloadURL,
		"expires_in":   int(h.PresignTTL.Seconds()),
	})
}
