package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
)

// allowedImageTypes are the only image MIME types accepted for
// /api/v1/photos/preset and as image members of /api/v1/documents/merge-pdf.
// HEIC/HEIF is explicitly out of scope per the project spec and is
// rejected with 400 even though a client may send it.
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

const pdfContentType = "application/pdf"

// detectContentType sniffs the real content type from the file bytes
// (magic number based), ignoring whatever Content-Type the client claimed
// in the multipart part — clients can lie about that header.
func detectContentType(data []byte) string {
	n := len(data)
	if n > 512 {
		n = 512
	}
	ct := http.DetectContentType(data[:n])
	// http.DetectContentType returns "image/webp" for WEBP and
	// "image/jpeg"/"image/png" for those formats already; for PDFs it
	// returns "application/pdf". No extra normalization needed today, but
	// keep this as a single seam in case that ever changes.
	return ct
}

func isAllowedImageType(contentType string) bool {
	return allowedImageTypes[contentType]
}

func isAllowedMergeInputType(contentType string) bool {
	return allowedImageTypes[contentType] || contentType == pdfContentType
}

// isMaxBytesError reports whether err came from a request body exceeding
// the http.MaxBytesReader limit installed in main.go / the handler.
func isMaxBytesError(err error) bool {
	var mbErr *http.MaxBytesError
	return errors.As(err, &mbErr)
}

// randomID returns a short random hex string suitable for use in generated
// output filenames/object keys.
func randomID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand.Read failing is effectively unrecoverable on any
		// supported platform; fall back to a fixed marker rather than
		// panicking mid-request.
		return "00000000"
	}
	return hex.EncodeToString(b)
}

// parsePositiveIntForm reads a form field as a positive int, returning
// (0, false) when it's missing/invalid/non-positive.
func parsePositiveIntForm(raw string) (int, bool) {
	if raw == "" {
		return 0, false
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return 0, false
	}
	return v, true
}
