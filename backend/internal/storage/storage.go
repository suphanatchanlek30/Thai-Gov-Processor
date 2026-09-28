// Package storage provides an S3-compatible object storage client used to
// hold processed files (photos, merged PDFs) behind presigned download
// URLs. It is the seam between local development (MinIO via S3_ENDPOINT)
// and real AWS S3 in later phases: when S3_ENDPOINT is empty the client
// falls back to the default AWS credential chain (IAM role, env vars,
// shared config, ...) exactly like a production deployment would use.
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

// Client is the storage seam the rest of the backend depends on. Handlers
// and processors should take this interface, not *S3Client, so tests can
// fake it.
type Client interface {
	// Put uploads data under key with the given content type.
	Put(ctx context.Context, key string, data []byte, contentType string) error
	// PresignGet returns a time-limited download URL for key.
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
}

// Config holds the S3-compatible connection settings, normally sourced
// from environment variables (see ConfigFromEnv).
type Config struct {
	Endpoint       string // custom endpoint (e.g. http://localhost:9000 for MinIO); empty = real AWS
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	ForcePathStyle bool
}

// ConfigFromEnv reads S3_ENDPOINT, S3_REGION, S3_BUCKET, S3_ACCESS_KEY,
// S3_SECRET_KEY, S3_FORCE_PATH_STYLE from the environment.
func ConfigFromEnv() Config {
	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "ap-southeast-1"
	}
	forcePathStyle, _ := strconv.ParseBool(os.Getenv("S3_FORCE_PATH_STYLE"))
	return Config{
		Endpoint:       os.Getenv("S3_ENDPOINT"),
		Region:         region,
		Bucket:         os.Getenv("S3_BUCKET"),
		AccessKey:      os.Getenv("S3_ACCESS_KEY"),
		SecretKey:      os.Getenv("S3_SECRET_KEY"),
		ForcePathStyle: forcePathStyle,
	}
}

// S3Client is the AWS SDK v2-backed implementation of Client.
type S3Client struct {
	api    *s3.Client
	presig *s3.PresignClient
	bucket string
}

// New builds an S3Client from cfg.
//
// When cfg.Endpoint is empty, credentials come from the default AWS
// credential chain (env vars, shared config/credentials files, EC2/ECS
// instance role, ...) — this is the path used against real S3 in later
// phases. When cfg.Endpoint is set (local MinIO), static credentials from
// cfg.AccessKey/cfg.SecretKey are used and requests are pointed at that
// custom endpoint with path-style addressing controlled by
// cfg.ForcePathStyle.
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

	return &S3Client{
		api:    api,
		presig: s3.NewPresignClient(api),
		bucket: cfg.Bucket,
	}, nil
}

// Put uploads data under key with the given content type.
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

// PresignGet returns a time-limited download URL for key.
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
