package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/service"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/storage"
)

type Handler struct {
	PhotoService    service.PhotoService
	DocumentService service.DocumentService
	Storage         storage.Client
	PresignTTL      time.Duration
	MaxUploadMB     int64
}

func New(photoService service.PhotoService, documentService service.DocumentService, store storage.Client, presignTTL time.Duration, maxUploadMB int64) *Handler {
	return &Handler{
		PhotoService:    photoService,
		DocumentService: documentService,
		Storage:         store,
		PresignTTL:      presignTTL,
		MaxUploadMB:     maxUploadMB,
	}
}

func (h *Handler) Register(r gin.IRouter) {
	r.GET("/healthz", h.Healthz)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/presets", h.Presets)
		v1.GET("/selftest", h.Selftest)
		v1.POST("/photos/preset", h.PhotoPreset)
		v1.POST("/documents/merge-pdf", h.MergePDF)
	}
}
