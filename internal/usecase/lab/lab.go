package lab

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"labshop/internal/usecase/port"
)

var (
	// ErrUnauthorized is returned when an invalid admin token is supplied.
	ErrUnauthorized = errors.New("unauthorized: invalid admin token")

	// ErrInvalidMode is returned when an unsupported cache mode is requested.
	ErrInvalidMode = errors.New("invalid cache mode")

	// ErrInvalidStatusCode is returned when a status code outside the allowed set is requested.
	ErrInvalidStatusCode = errors.New("invalid status code")

	// ErrSlowBusy is returned when more than 20 concurrent slow requests are active.
	ErrSlowBusy = errors.New("server busy: too many slow requests")

	// ErrUploadTooLarge is returned when an uploaded file exceeds the configured size limit.
	ErrUploadTooLarge = errors.New("upload size exceeds limit")
)

// AllowedStatusCodes defines the strictly permitted HTTP status codes for /api/status/{code}.
var AllowedStatusCodes = map[int]bool{
	200: true, 204: true, 301: true, 302: true, 304: true,
	400: true, 401: true, 403: true, 404: true, 429: true,
	500: true, 502: true, 503: true, 504: true,
}

var validCacheModes = map[string]bool{
	"public":     true,
	"no-store":   true,
	"etag":       true,
	"vary":       true,
	"set-cookie": true,
}

var unsafeFilenamePattern = regexp.MustCompile(`[^a-zA-Z0-9._\-]`)

// Config configures limits and tokens for lab operations.
type Config struct {
	AdminToken   string
	MaxSlowMS    int
	MaxPayloadKB int
	MaxUploadMB  int
}

// TimeResult holds the result of a time request.
type TimeResult struct {
	GeneratedAt time.Time
	RequestID   string
}

// CacheResult holds the result of a cache test request.
type CacheResult struct {
	Mode        string
	GeneratedAt time.Time
	RequestID   string
	Version     string
	ETag        string
}

// UploadResult details an uploaded file without persisting to disk.
type UploadResult struct {
	SHA256      string
	Size        int64
	Filename    string
	ContentType string
}

// Usecase handles application logic for testing CDN, caching, and WAF behaviors.
type Usecase struct {
	counter port.HitCounter
	clock   port.Clock
	cfg     Config
	slowSem chan struct{}
}

// New constructs a new Lab usecase instance.
func New(counter port.HitCounter, clock port.Clock, cfg Config) *Usecase {
	if cfg.MaxSlowMS <= 0 {
		cfg.MaxSlowMS = 60000
	}
	if cfg.MaxPayloadKB <= 0 {
		cfg.MaxPayloadKB = 5120
	}
	if cfg.MaxUploadMB <= 0 {
		cfg.MaxUploadMB = 10
	}
	if clock == nil {
		clock = port.RealClock{}
	}

	return &Usecase{
		counter: counter,
		clock:   clock,
		cfg:     cfg,
		slowSem: make(chan struct{}, 20), // Max 20 concurrent slow requests across server
	}
}

// GetTime returns the current timestamp and request ID.
func (u *Usecase) GetTime(requestID string) TimeResult {
	return TimeResult{
		GeneratedAt: u.clock.Now(),
		RequestID:   requestID,
	}
}

// GetCache validates the mode and returns cache test metadata.
func (u *Usecase) GetCache(mode, version, requestID string) (CacheResult, error) {
	if !validCacheModes[mode] {
		return CacheResult{}, ErrInvalidMode
	}

	h := sha256.New()
	h.Write([]byte("lab-etag-" + version))
	etag := fmt.Sprintf("\"%s\"", hex.EncodeToString(h.Sum(nil)))

	return CacheResult{
		Mode:        mode,
		GeneratedAt: u.clock.Now(),
		RequestID:   requestID,
		Version:     version,
		ETag:        etag,
	}, nil
}

// GetStats returns current origin hit counts per path.
func (u *Usecase) GetStats() map[string]int64 {
	return u.counter.GetStats()
}

// ResetStats resets hit counts if the admin token matches.
func (u *Usecase) ResetStats(token string) error {
	if token == "" || token != u.cfg.AdminToken {
		return ErrUnauthorized
	}
	u.counter.Reset()
	return nil
}

// Slow pauses execution for the requested duration, clamped to MaxSlowMS, enforcing max 20 concurrency.
func (u *Usecase) Slow(ctx context.Context, requestedMS int) (int, error) {
	if requestedMS <= 0 {
		requestedMS = 1000
	}
	if requestedMS > u.cfg.MaxSlowMS {
		requestedMS = u.cfg.MaxSlowMS
	}

	select {
	case u.slowSem <- struct{}{}:
		defer func() { <-u.slowSem }()
	default:
		return 0, ErrSlowBusy
	}

	timer := time.NewTimer(time.Duration(requestedMS) * time.Millisecond)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-timer.C:
		return requestedMS, nil
	}
}

// GeneratePayload generates compressible repetitive data of the requested size in KB.
func (u *Usecase) GeneratePayload(requestedKB int) (string, error) {
	if requestedKB <= 0 {
		requestedKB = 1
	}
	if requestedKB > u.cfg.MaxPayloadKB {
		requestedKB = u.cfg.MaxPayloadKB
	}

	totalBytes := requestedKB * 1024
	pattern := "LabShop-CDN-Compression-Test-Payload-Data-Block-ABCXYZ1234567890- "
	var sb strings.Builder
	sb.Grow(totalBytes)

	for sb.Len() < totalBytes {
		remaining := totalBytes - sb.Len()
		if remaining < len(pattern) {
			sb.WriteString(pattern[:remaining])
			break
		}
		sb.WriteString(pattern)
	}

	return sb.String(), nil
}

// IsAllowedStatusCode checks if the given HTTP status code is supported by the status endpoint.
func (u *Usecase) IsAllowedStatusCode(code int) bool {
	return AllowedStatusCodes[code]
}

// ProcessUpload processes an uploaded file stream, calculates its SHA-256 and size,
// sanitizes the filename, and validates against MaxUploadMB without saving to disk.
func (u *Usecase) ProcessUpload(filename, contentType string, reader io.Reader) (UploadResult, error) {
	maxBytes := int64(u.cfg.MaxUploadMB) * 1024 * 1024

	cleanFilename := sanitizeFilename(filename)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	hasher := sha256.New()
	// Read at most maxBytes + 1 to detect overflow
	written, err := io.Copy(hasher, io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return UploadResult{}, err
	}

	if written > maxBytes {
		return UploadResult{}, ErrUploadTooLarge
	}

	return UploadResult{
		SHA256:      hex.EncodeToString(hasher.Sum(nil)),
		Size:        written,
		Filename:    cleanFilename,
		ContentType: contentType,
	}, nil
}

func sanitizeFilename(name string) string {
	base := filepath.Base(name)
	base = strings.ReplaceAll(base, `\`, "")
	base = strings.ReplaceAll(base, `/`, "")
	sanitized := unsafeFilenamePattern.ReplaceAllString(base, "_")
	if sanitized == "" || sanitized == "." || sanitized == ".." {
		sanitized = "upload.bin"
	}
	return sanitized
}
