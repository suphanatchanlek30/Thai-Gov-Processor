// Package storage provides an S3-compatible object storage client. Empty
// S3_ENDPOINT uses the default AWS credential chain (real S3, later
// phases); a set S3_ENDPOINT (local MinIO) uses static credentials against
// that endpoint with path-style addressing.
package storage

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client interface {
	Put(ctx context.Context, key string, data []byte, contentType string) error
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
}

type Config struct {
	Endpoint       string // custom endpoint (e.g. MinIO); empty = real AWS
	PublicEndpoint string // endpoint to sign presigned URLs against, if different from Endpoint (e.g. MinIO reachable at a different host/port from outside Docker)
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	ForcePathStyle bool
}

func ConfigFromEnv() Config {
	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "ap-southeast-1"
	}
	forcePathStyle, _ := strconv.ParseBool(os.Getenv("S3_FORCE_PATH_STYLE"))
	endpoint := os.Getenv("S3_ENDPOINT")
	publicEndpoint := os.Getenv("S3_PUBLIC_ENDPOINT")
	if publicEndpoint == "" {
		publicEndpoint = endpoint
	}
	return Config{
		Endpoint:       endpoint,
		PublicEndpoint: publicEndpoint,
		Region:         region,
		Bucket:         os.Getenv("S3_BUCKET"),
		AccessKey:      os.Getenv("S3_ACCESS_KEY"),
		SecretKey:      os.Getenv("S3_SECRET_KEY"),
		ForcePathStyle: forcePathStyle,
	}
}

type S3Client struct {
	api    *s3.Client
	presig *s3.PresignClient
	bucket string
}

func New(ctx context.Context, cfg Config) (*S3Client, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("storage: S3_BUCKET is required")
	}

	var optFns []func(*awsconfig.LoadOptions) error
	optFns = append(optFns, awsconfig.WithRegion(cfg.Region))
	if cfg.Endpoint != "" {
		optFns = append(optFns, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, fmt.Errorf("storage: load aws config: %w", err)
	}

	api := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.ForcePathStyle
	})

	// Presigned URLs are handed to callers outside our network (e.g. a
	// browser), so they must be signed against PublicEndpoint rather than
	// the internal Endpoint used for server-side Put/Get calls when the two
	// differ (e.g. MinIO reachable as "minio:9000" from the backend
	// container but only as "localhost:9000" from the host).
	presignAPI := api
	if cfg.PublicEndpoint != cfg.Endpoint {
		presignAPI = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			if cfg.PublicEndpoint != "" {
				o.BaseEndpoint = aws.String(cfg.PublicEndpoint)
			}
			o.UsePathStyle = cfg.ForcePathStyle
		})
	}

	return &S3Client{
		api:    api,
		presig: s3.NewPresignClient(presignAPI),
		bucket: cfg.Bucket,
	}, nil
}

func (c *S3Client) Put(ctx context.Context, key string, data []byte, contentType string) error {
	_, err := c.api.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("storage: put %q: %w", key, err)
	}
	return nil
}

func (c *S3Client) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	req, err := c.presig.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("storage: presign %q: %w", key, err)
	}
	return req.URL, nil
}
