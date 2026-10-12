package lab

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"testing"
	"time"

	"labshop/internal/usecase/port"
)

type fixedClock struct {
	t time.Time
}

func (f fixedClock) Now() time.Time {
	return f.t
}

type mockCounter struct {
	sync.RWMutex
	hits map[string]int64
}

func newMockCounter() *mockCounter {
	return &mockCounter{hits: make(map[string]int64)}
}

var _ port.HitCounter = (*mockCounter)(nil)

func (m *mockCounter) Increment(path string) {
	m.Lock()
	defer m.Unlock()
	m.hits[path]++
}

func (m *mockCounter) GetStats() map[string]int64 {
	m.RLock()
	defer m.RUnlock()
	cpy := make(map[string]int64, len(m.hits))
	for k, v := range m.hits {
		cpy[k] = v
	}
	return cpy
}

func (m *mockCounter) Reset() {
	m.Lock()
	defer m.Unlock()
	m.hits = make(map[string]int64)
}

func setupUsecase() (*Usecase, *mockCounter, time.Time) {
	fixedTime := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	counter := newMockCounter()
	uc := New(counter, fixedClock{t: fixedTime}, Config{
		AdminToken:   "super-secret-admin-token",
		MaxSlowMS:    5000,
		MaxPayloadKB: 100,
		MaxUploadMB:  1,
	})
	return uc, counter, fixedTime
}

func TestGetTime(t *testing.T) {
	uc, _, fixedTime := setupUsecase()
	res := uc.GetTime("req-1234")
	if !res.GeneratedAt.Equal(fixedTime) {
		t.Fatalf("expected generated_at %v, got %v", fixedTime, res.GeneratedAt)
	}
	if res.RequestID != "req-1234" {
		t.Fatalf("expected request_id req-1234, got %s", res.RequestID)
	}
}

func TestGetCache(t *testing.T) {
	uc, _, fixedTime := setupUsecase()

	validModes := []string{"public", "no-store", "etag", "vary", "set-cookie"}
	for _, mode := range validModes {
		res, err := uc.GetCache(mode, "v1", "req-1")
		if err != nil {
			t.Fatalf("unexpected error for mode %s: %v", mode, err)
		}
		if res.Mode != mode {
			t.Fatalf("expected mode %s, got %s", mode, res.Mode)
		}
		if !res.GeneratedAt.Equal(fixedTime) {
			t.Fatalf("expected time %v, got %v", fixedTime, res.GeneratedAt)
		}
		if res.RequestID != "req-1" {
			t.Fatalf("expected req-1, got %s", res.RequestID)
		}
		if res.Version != "v1" {
			t.Fatalf("expected v1, got %s", res.Version)
		}
		if res.ETag == "" {
			t.Fatalf("expected non-empty etag")
		}
	}

	_, err := uc.GetCache("unsupported", "v1", "req-1")
	if !errors.Is(err, ErrInvalidMode) {
		t.Fatalf("expected ErrInvalidMode, got %v", err)
	}
}

func TestGetStatsAndReset(t *testing.T) {
	uc, counter, _ := setupUsecase()

	counter.Increment("/api/health")
	counter.Increment("/api/health")
	counter.Increment("/api/time")

	stats := uc.GetStats()
	if stats["/api/health"] != 2 {
		t.Fatalf("expected 2 health hits, got %d", stats["/api/health"])
	}
	if stats["/api/time"] != 1 {
		t.Fatalf("expected 1 time hit, got %d", stats["/api/time"])
	}

	// Reset with wrong token
	err := uc.ResetStats("wrong-token")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
	if uc.GetStats()["/api/health"] != 2 {
		t.Fatalf("stats should not have been cleared with invalid token")
	}

	// Reset with empty token
	err = uc.ResetStats("")
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}

	// Reset with correct token
	err = uc.ResetStats("super-secret-admin-token")
	if err != nil {
		t.Fatalf("unexpected error resetting stats: %v", err)
	}
	if len(uc.GetStats()) != 0 {
		t.Fatalf("expected stats to be empty after reset, got %v", uc.GetStats())
	}
}

func TestSlow(t *testing.T) {
	uc, _, _ := setupUsecase()

	// Short sleep
	slept, err := uc.Slow(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error in Slow: %v", err)
	}
	if slept != 10 {
		t.Fatalf("expected 10 ms slept, got %d", slept)
	}

	// Context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = uc.Slow(ctx, 1000)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	// Concurrency limiter test: saturate 20 workers
	var wg sync.WaitGroup
	started := make(chan struct{})
	release := make(chan struct{})

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case uc.slowSem <- struct{}{}:
				started <- struct{}{}
				<-release
				<-uc.slowSem
			default:
			}
		}()
	}

	for i := 0; i < 20; i++ {
		<-started
	}

	// 21st request should be rejected immediately with ErrSlowBusy
	_, busyErr := uc.Slow(context.Background(), 10)
	if !errors.Is(busyErr, ErrSlowBusy) {
		t.Fatalf("expected ErrSlowBusy, got %v", busyErr)
	}

	close(release)
	wg.Wait()
}

func TestGeneratePayload(t *testing.T) {
	uc, _, _ := setupUsecase()

	// 2 KB
	payload, err := uc.GeneratePayload(2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payload) != 2*1024 {
		t.Fatalf("expected 2048 bytes, got %d", len(payload))
	}

	// Clamped to MaxPayloadKB (100 KB)
	payloadClamped, err := uc.GeneratePayload(500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payloadClamped) != 100*1024 {
		t.Fatalf("expected 102400 bytes, got %d", len(payloadClamped))
	}

	// Zero defaults to 1 KB
	payloadMin, err := uc.GeneratePayload(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payloadMin) != 1*1024 {
		t.Fatalf("expected 1024 bytes, got %d", len(payloadMin))
	}
}

func TestIsAllowedStatusCode(t *testing.T) {
	uc, _, _ := setupUsecase()

	allowed := []int{200, 204, 301, 302, 304, 400, 401, 403, 404, 429, 500, 502, 503, 504}
	for _, code := range allowed {
		if !uc.IsAllowedStatusCode(code) {
			t.Fatalf("code %d should be allowed", code)
		}
	}

	disallowed := []int{100, 201, 202, 307, 308, 405, 418, 501}
	for _, code := range disallowed {
		if uc.IsAllowedStatusCode(code) {
			t.Fatalf("code %d should not be allowed", code)
		}
	}
}

func TestProcessUpload(t *testing.T) {
	uc, _, _ := setupUsecase()

	content := []byte("Hello, LabShop upload testing payload!")
	h := sha256.New()
	h.Write(content)
	expectedSHA := hex.EncodeToString(h.Sum(nil))

	res, err := uc.ProcessUpload("../../../unsafe/evil.sh", "text/plain", bytes.NewReader(content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.SHA256 != expectedSHA {
		t.Fatalf("expected SHA %s, got %s", expectedSHA, res.SHA256)
	}
	if res.Size != int64(len(content)) {
		t.Fatalf("expected size %d, got %d", len(content), res.Size)
	}
	if res.Filename != "evil.sh" {
		t.Fatalf("expected evil.sh, got %s", res.Filename)
	}
	if res.ContentType != "text/plain" {
		t.Fatalf("expected text/plain, got %s", res.ContentType)
	}

	// Exceed limit (MaxUploadMB = 1MB)
	hugeContent := make([]byte, 1024*1024+10)
	_, err = uc.ProcessUpload("large.bin", "application/octet-stream", bytes.NewReader(hugeContent))
	if !errors.Is(err, ErrUploadTooLarge) {
		t.Fatalf("expected ErrUploadTooLarge, got %v", err)
	}
}
