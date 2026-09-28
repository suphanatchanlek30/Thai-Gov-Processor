package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
)

// HEIC/HEIF is explicitly out of scope and rejected with 400 even if a client sends it.
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

const pdfContentType = "application/pdf"

// Sniffs the real type from file bytes instead of trusting the client's declared Content-Type.
func detectContentType(data []byte) string {
	n := len(data)
	if n > 512 {
		n = 512
	}
	return http.DetectContentType(data[:n])
}

func isAllowedImageType(contentType string) bool {
	return allowedImageTypes[contentType]
}

func isAllowedMergeInputType(contentType string) bool {
	return allowedImageTypes[contentType] || contentType == pdfContentType
}

func isMaxBytesError(err error) bool {
	var mbErr *http.MaxBytesError
	return errors.As(err, &mbErr)
}

func randomID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "00000000" // avoid panicking mid-request
	}
	return hex.EncodeToString(b)
}

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
