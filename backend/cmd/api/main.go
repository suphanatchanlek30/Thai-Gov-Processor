// Command api is the Thai Gov Photo & Doc Processor backend entry point.
//
// This is the Phase 0 scaffold: routing, validation, storage and the
// PhotoProcessor/PDFMerger contracts are all real, but the processors
// themselves are temporary stubs (internal/processor/stub) that return 501
// until a follow-up PR wires in the real image/PDF processing.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/handler"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/middleware"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/processor/stub"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/storage"
)

func main() {
	port := envOr("PORT", "8080")
	maxUploadMB := envInt64Or("MAX_UPLOAD_MB", 15)
	presignTTL := time.Duration(envInt64Or("PRESIGN_TTL_SECONDS", 3600)) * time.Second

	ctx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	storageClient, err := storage.New(ctx, storage.ConfigFromEnv())
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	// --- Processor wiring -------------------------------------------------
	// These two lines are the entire seam for the follow-up
	// feat/image-pdf-processor PR: swap stub.NewNotImplementedPhotoProcessor
	// / stub.NewNotImplementedPDFMerger for the real constructors and
	// nothing else in this file (or in internal/handler) needs to change.
	photoProcessor := stub.NewNotImplementedPhotoProcessor()
	pdfMerger := stub.NewNotImplementedPDFMerger()
	// ------------------------------------------------------------------

	h := handler.New(photoProcessor, pdfMerger, storageClient, presignTTL, maxUploadMB)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(middleware.Recovery(), middleware.Logger())
	h.Register(router)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	stopSignals()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	log.Println("shutdown complete")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt64Or(key string, fallback int64) int64 {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}
	return v
}
