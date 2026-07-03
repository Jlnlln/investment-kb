package ai

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCompleteJSONRetriesTruncatedJSONThenSucceeds(t *testing.T) {
	restoreBackoffs := disableCompleteJSONBackoffsForTest()
	defer restoreBackoffs()

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := atomic.AddInt32(&calls, 1)
		if call == 1 {
			writeMessagesResponse(t, w, `{"title":"安全边际"`)
			return
		}
		writeMessagesResponse(t, w, `{"title":"安全边际"}`)
	}))
	defer server.Close()

	client := newCustomClient(&Config{
		Model:       "test-model",
		BaseURL:     server.URL,
		APIKey:      "test-key",
		TimeoutSec:  5,
		Temperature: 0,
	})

	debugDir := t.TempDir()
	ctx := WithDebugInfo(context.Background(), `testdata\inputs\rule_safety_margin.md`, debugDir)
	var got map[string]string
	if err := client.CompleteJSON(ctx, "system", "user", &got); err != nil {
		t.Fatalf("CompleteJSON returned error: %v", err)
	}
	if got["title"] != "安全边际" {
		t.Fatalf("title = %q, want 安全边际", got["title"])
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	assertDebugFileCount(t, debugDir, 1)
}

func TestCompleteJSONReturnsErrorAfterThreeJSONFailures(t *testing.T) {
	restoreBackoffs := disableCompleteJSONBackoffsForTest()
	defer restoreBackoffs()

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		writeMessagesResponse(t, w, `{"title":"截断"`)
	}))
	defer server.Close()

	client := newCustomClient(&Config{
		Model:      "test-model",
		BaseURL:    server.URL,
		APIKey:     "test-key",
		TimeoutSec: 5,
	})

	debugDir := t.TempDir()
	ctx := WithDebugInfo(context.Background(), "input.md", debugDir)
	var got map[string]string
	err := client.CompleteJSON(ctx, "system", "user", &got)
	if err == nil {
		t.Fatal("CompleteJSON expected error, got nil")
	}
	if !strings.Contains(err.Error(), "JSON 解析失败") {
		t.Fatalf("error = %v, want JSON parse error", err)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3", calls)
	}
	assertDebugFileCount(t, debugDir, 3)
}

func TestCompleteJSONDoesNotRetryNonRetryableHTTPError(t *testing.T) {
	restoreBackoffs := disableCompleteJSONBackoffsForTest()
	defer restoreBackoffs()

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		http.Error(w, `{"message":"bad request"}`, http.StatusBadRequest)
	}))
	defer server.Close()

	client := newCustomClient(&Config{
		Model:      "test-model",
		BaseURL:    server.URL,
		APIKey:     "test-key",
		TimeoutSec: 5,
	})

	var got map[string]string
	err := client.CompleteJSON(context.Background(), "system", "user", &got)
	if err == nil {
		t.Fatal("CompleteJSON expected error, got nil")
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
}

func TestCompleteJSONRetriesHTTP429ThenSucceeds(t *testing.T) {
	restoreBackoffs := disableCompleteJSONBackoffsForTest()
	defer restoreBackoffs()

	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := atomic.AddInt32(&calls, 1)
		if call == 1 {
			w.Header().Set("Retry-After", "0")
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}
		writeMessagesResponse(t, w, `{"title":"限流后成功"}`)
	}))
	defer server.Close()

	client := newCustomClient(&Config{
		Model:      "test-model",
		BaseURL:    server.URL,
		APIKey:     "test-key",
		TimeoutSec: 5,
	})

	var got map[string]string
	if err := client.CompleteJSON(context.Background(), "system", "user", &got); err != nil {
		t.Fatalf("CompleteJSON returned error: %v", err)
	}
	if got["title"] != "限流后成功" {
		t.Fatalf("title = %q, want 限流后成功", got["title"])
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestCompleteJSONDebugFileIncludesInputAttemptErrorAndRawResponse(t *testing.T) {
	restoreBackoffs := disableCompleteJSONBackoffsForTest()
	defer restoreBackoffs()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeMessagesResponse(t, w, `{"title":"半截"`)
	}))
	defer server.Close()

	client := newCustomClient(&Config{
		Model:      "test-model",
		BaseURL:    server.URL,
		APIKey:     "test-key",
		TimeoutSec: 5,
	})

	debugDir := t.TempDir()
	ctx := WithDebugInfo(context.Background(), `G:\GoCode\investment-kb\testdata\inputs\rule_safety_margin.md`, debugDir)
	var got map[string]string
	_ = client.CompleteJSON(ctx, "system", "user", &got)

	files, err := os.ReadDir(debugDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected debug files, got none")
	}
	data, err := os.ReadFile(filepath.Join(debugDir, files[0].Name()))
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	content := string(data)
	for _, want := range []string{
		`input path: G:\GoCode\investment-kb\testdata\inputs\rule_safety_margin.md`,
		"attempt: 1",
		"error: JSON 解析失败",
		`raw AI response:`,
		`{"title":"半截"`,
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("debug content missing %q:\n%s", want, content)
		}
	}
}

func writeMessagesResponse(t *testing.T, w http.ResponseWriter, text string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	_, err := fmt.Fprintf(w, `{"content":[{"type":"text","text":%q}]}`, text)
	if err != nil {
		t.Fatalf("write response failed: %v", err)
	}
}

func disableCompleteJSONBackoffsForTest() func() {
	oldBackoffs := completeJSONBackoffs
	oldFallback := retryAfterFallback
	oldLastCallTime := lastCallTime
	completeJSONBackoffs = []time.Duration{0, 0}
	retryAfterFallback = 0
	lastCallTime = time.Time{}
	return func() {
		completeJSONBackoffs = oldBackoffs
		retryAfterFallback = oldFallback
		lastCallTime = oldLastCallTime
	}
}

func assertDebugFileCount(t *testing.T, dir string, want int) {
	t.Helper()
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(files) != want {
		t.Fatalf("debug file count = %d, want %d", len(files), want)
	}
}
