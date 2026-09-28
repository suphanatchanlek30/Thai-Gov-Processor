package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/apierr"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/preset"
)

// PhotoPreset handles POST /api/v1/photos/preset: validates the upload and
// the chosen preset, runs it through the PhotoService, stores the result,
// and returns a presigned download URL.
//
// While PhotoService is the stub implementation this always returns 501
// after passing validation — validation itself is fully enforced today.
func (h *Handler) PhotoPreset(c *gin.Context) {
	maxBytes := h.MaxUploadMB * 1024 * 1024
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)

	presetName := c.PostForm("preset")
	if presetName == "" {
		apierr.BadRequest(c, "preset is required")
		return
	}
	if !preset.IsValidName(presetName) {
		apierr.BadRequest(c, "preset must be one of: ocsc, passport, teacher, custom")
		return
	}

	var custom *preset.Custom
	if presetName == preset.CustomName {
		width, ok := parsePositiveIntForm(c.PostForm("width"))
		if !ok {
			apierr.BadRequest(c, "width is required and must be a positive integer for preset=custom")
			return
		}
		height, ok := parsePositiveIntForm(c.PostForm("height"))
		if !ok {
			apierr.BadRequest(c, "height is required and must be a positive integer for preset=custom")
			return
		}
		maxKB, ok := parsePositiveIntForm(c.PostForm("max_kb"))
		if !ok {
			apierr.BadRequest(c, "max_kb is required and must be a positive integer for preset=custom")
			return
		}
		custom = &preset.Custom{Width: width, Height: height, MaxKB: maxKB}
		if err := preset.ValidateCustom(*custom); err != nil {
			apierr.BadRequest(c, err.Error())
			return
		}
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		if isMaxBytesError(err) {
			apierr.PayloadTooLarge(c, fmt.Sprintf("file exceeds maximum upload size of %d MB", h.MaxUploadMB))
			return
		}
		apierr.BadRequest(c, "file is required")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		apierr.BadRequest(c, "could not read uploaded file")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		if isMaxBytesError(err) {
			apierr.PayloadTooLarge(c, fmt.Sprintf("file exceeds maximum upload size of %d MB", h.MaxUploadMB))
			return
		}
		apierr.BadRequest(c, "could not read uploaded file")
		return
	}

	contentType := detectContentType(data)
	if !isAllowedImageType(contentType) {
		apierr.BadRequest(c, fmt.Sprintf("unsupported file type %q: only JPG, PNG and WEBP are supported (HEIC is not supported)", contentType))
		return
	}

	result, err := h.PhotoService.ProcessPreset(data, presetName, custom)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	filename := fmt.Sprintf("%s_%s.jpg", presetName, randomID())
	key := "processed/" + filename

	ctx := c.Request.Context()
	if err := h.Storage.Put(ctx, key, result.Data, "image/jpeg"); err != nil {
		apierr.Internal(c, "failed to store processed file")
		return
	}
	downloadURL, err := h.Storage.PresignGet(ctx, key, h.PresignTTL)
	if err != nil {
		apierr.Internal(c, "failed to generate download URL")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"filename":     filename,
		"width":        result.Width,
		"height":       result.Height,
		"size_kb":      result.SizeKB,
		"quality":      result.Quality,
		"download_url": downloadURL,
		"expires_in":   int(h.PresignTTL.Seconds()),
	})
}
