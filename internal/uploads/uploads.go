package uploads

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrFileTooLarge    = errors.New("file exceeds maximum size")
	ErrInvalidMimeType = errors.New("file type not allowed")
)

// DefaultMaxFileSize is 5MB
const DefaultMaxFileSize = 5 * 1024 * 1024

// AllowedMimeTypes lists the allowed image MIME types
var AllowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// Config holds upload service configuration
type Config struct {
	UploadsDir  string
	MaxFileSize int64
}

// Service handles file uploads
type Service struct {
	config Config
}

// NewService creates a new upload service
func NewService(config Config) (*Service, error) {
	// Set defaults
	if config.UploadsDir == "" {
		config.UploadsDir = "./uploads"
	}
	if config.MaxFileSize == 0 {
		config.MaxFileSize = DefaultMaxFileSize
	}

	// Ensure uploads directory exists
	if err := os.MkdirAll(config.UploadsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create uploads directory: %w", err)
	}

	return &Service{config: config}, nil
}

// UploadResult contains information about an uploaded file
type UploadResult struct {
	StoredName string
	MimeType   string
	SizeBytes  int64
}

// SaveFile saves an uploaded file and returns the stored filename
func (s *Service) SaveFile(file *multipart.FileHeader) (*UploadResult, error) {
	// Check file size
	if file.Size > s.config.MaxFileSize {
		return nil, ErrFileTooLarge
	}

	// Check MIME type
	mimeType := file.Header.Get("Content-Type")
	if !AllowedMimeTypes[mimeType] {
		return nil, ErrInvalidMimeType
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Generate UUID-based filename with original extension
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		// Infer extension from mime type
		switch mimeType {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		}
	}
	storedName := uuid.New().String() + strings.ToLower(ext)

	// Create destination file
	dstPath := filepath.Join(s.config.UploadsDir, storedName)
	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file contents
	written, err := io.Copy(dst, src)
	if err != nil {
		// Clean up on error
		os.Remove(dstPath)
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &UploadResult{
		StoredName: storedName,
		MimeType:   mimeType,
		SizeBytes:  written,
	}, nil
}

// GetFilePath returns the full path to a stored file
func (s *Service) GetFilePath(storedName string) string {
	return filepath.Join(s.config.UploadsDir, storedName)
}

// DeleteFile removes a stored file
func (s *Service) DeleteFile(storedName string) error {
	path := filepath.Join(s.config.UploadsDir, storedName)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// FileExists checks if a file exists
func (s *Service) FileExists(storedName string) bool {
	path := filepath.Join(s.config.UploadsDir, storedName)
	_, err := os.Stat(path)
	return err == nil
}
