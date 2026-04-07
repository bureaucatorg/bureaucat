package uploads

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

var (
	ErrFileTooLarge = errors.New("file exceeds maximum size")
)

// DefaultMaxFileSize is 10MB
const DefaultMaxFileSize = 10 * 1024 * 1024

// Config holds upload service configuration.
type Config struct {
	S3Endpoint  string
	BucketName  string
	AccessKeyID string
	SecretKey   string
	UseSSL      bool
	MaxFileSize int64
}

// Service handles file uploads to S3-compatible storage.
type Service struct {
	client     *s3.Client
	bucketName string
	maxFileSize int64
}

// NewService creates a new S3-backed upload service.
func NewService(cfg Config) (*Service, error) {
	if cfg.MaxFileSize == 0 {
		cfg.MaxFileSize = DefaultMaxFileSize
	}

	// Build S3 client with static credentials and custom endpoint.
	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(cfg.S3Endpoint),
		Region:       "garage",
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretKey, ""),
		UsePathStyle: true,
	})

	// Verify the bucket exists.
	_, err := client.HeadBucket(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(cfg.BucketName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to access S3 bucket %q: %w", cfg.BucketName, err)
	}

	return &Service{
		client:      client,
		bucketName:  cfg.BucketName,
		maxFileSize: cfg.MaxFileSize,
	}, nil
}

// UploadResult contains information about an uploaded file.
type UploadResult struct {
	StoredName string
	MimeType   string
	SizeBytes  int64
}

// SaveFile uploads a multipart file to S3 and returns the result.
func (s *Service) SaveFile(file *multipart.FileHeader) (*UploadResult, error) {
	if file.Size > s.maxFileSize {
		return nil, ErrFileTooLarge
	}

	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Generate UUID-based key with original extension.
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = extFromMime(mimeType)
	}
	storedName := uuid.New().String() + strings.ToLower(ext)

	_, err = s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        aws.String(s.bucketName),
		Key:           aws.String(storedName),
		Body:          src,
		ContentLength: aws.Int64(file.Size),
		ContentType:   aws.String(mimeType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	return &UploadResult{
		StoredName: storedName,
		MimeType:   mimeType,
		SizeBytes:  file.Size,
	}, nil
}

// GetFile returns a reader for the stored file from S3.
func (s *Service) GetFile(storedName string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(storedName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}
	return out.Body, nil
}

// DeleteFile removes a stored file from S3.
func (s *Service) DeleteFile(storedName string) error {
	_, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(storedName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}
	return nil
}

// FileExists checks if a file exists in S3.
func (s *Service) FileExists(storedName string) bool {
	_, err := s.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(storedName),
	})
	return err == nil
}

// extFromMime returns a file extension for a given MIME type.
func extFromMime(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	case "text/plain":
		return ".txt"
	case "application/msword":
		return ".doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return ".docx"
	case "application/vnd.ms-excel":
		return ".xls"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return ".xlsx"
	default:
		return ""
	}
}
