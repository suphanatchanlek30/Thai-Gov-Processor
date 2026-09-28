// Command api is the Thai Gov Photo & Doc Processor backend entry point.
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
	"github.com/joho/godotenv"

	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/handler"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/middleware"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/service/document"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/service/photo"
	"github.com/suphanatchanlek30/Thai-Gov-Processor/backend/internal/storage"
)

func main() {
	_ = godotenv.Load() // no-op if .env is absent (e.g. running inside docker-compose)

	port := envOr("PORT", "8080")
	maxUploadMB := envInt64Or("MAX_UPLOAD_MB", 15)
	presignTTL := time.Duration(envInt64Or("PRESIGN_TTL_SECONDS", 3600)) * time.Second

	ctx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	storageClient, err := storage.New(ctx, storage.ConfigFromEnv())
	if err != nil {
		log.Fatalf("storage: %v", err)
	}

	photoService := photo.NewService()
	documentService := document.NewService()

	h := handler.New(photoService, documentService, storageClient, presignTTL, maxUploadMB)

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
