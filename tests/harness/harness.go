package harness

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/heidericklucas/go-security-swe-bench/app/auth"
	"github.com/heidericklucas/go-security-swe-bench/app/clock"
	"github.com/heidericklucas/go-security-swe-bench/app/fetch"
	"github.com/heidericklucas/go-security-swe-bench/app/httpapi"
	"github.com/heidericklucas/go-security-swe-bench/app/seed"
	"github.com/heidericklucas/go-security-swe-bench/app/store"
)

// App is the test harness providing isolated test infrastructure.
type App struct {
	Base    string
	Fetcher *fetch.Fetcher
	Token   func(t *testing.T, userID int64, isAdmin bool) string
	Do      func(t *testing.T, method, path, bearer string, body any) *http.Response
	Data    seed.Data
	server  *httptest.Server
	store   *store.Store
	verify  *auth.Verifier
}

// Start spins up an isolated test app instance with seeded data.
func Start(t *testing.T) *App {
	t.Helper()

	// Fixed clock for determinism
	clk := clock.Fixed(time.Unix(1_700_000_000, 0))

	// Create in-memory database
	st, err := store.New(":memory:", clk)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Seed data. The attachments dir lives under a per-test root; the
	// path-traversal target (app_secret.txt) is written OUTSIDE it, here in the
	// held-out harness (never in candidate-editable app/seed), so a solver
	// cannot win by deleting the sentinel.
	root := t.TempDir()
	filesDir := filepath.Join(root, "attachments")
	data, err := seed.Load(st, filesDir)
	if err != nil {
		t.Fatalf("failed to seed data: %v", err)
	}
	sentinelToken := "TRAVERSAL_SENTINEL_TOKEN"
	if err := os.WriteFile(filepath.Join(root, "app_secret.txt"), []byte(sentinelToken), 0o600); err != nil {
		t.Fatalf("failed to write sentinel: %v", err)
	}
	data.SentinelToken = sentinelToken

	// Create verifier with fixed keys
	keys := seed.FixedRSAKeys()
	verifier := auth.NewVerifier(keys, clk)

	// Create injected fetcher
	f := fetch.New()

	// Start HTTP server
	handler := httpapi.New(st, verifier, f, filesDir)
	httpServer := httptest.NewServer(handler)
	t.Cleanup(func() { httpServer.Close() })

	// Create App instance
	app := &App{
		Base:    httpServer.URL,
		Fetcher: f,
		Data:    data,
		server:  httpServer,
		store:   st,
		verify:  verifier,
	}

	// Token func generates valid RS256 tokens
	app.Token = func(t *testing.T, userID int64, isAdmin bool) string {
		t.Helper()
		tok, err := verifier.Issue(userID, isAdmin, 1*time.Hour)
		if err != nil {
			t.Fatalf("failed to issue token: %v", err)
		}
		return tok
	}

	// Do func makes authenticated requests
	app.Do = func(t *testing.T, method, path, bearer string, bodyData any) *http.Response {
		t.Helper()
		var bodyReader io.Reader
		if bodyData != nil {
			buf := &bytes.Buffer{}
			if err := json.NewEncoder(buf).Encode(bodyData); err != nil {
				t.Fatalf("failed to encode body: %v", err)
			}
			bodyReader = buf
		}
		req, err := http.NewRequestWithContext(context.Background(), method, app.Base+path, bodyReader)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		return resp
	}

	return app
}

// ReadBody reads and returns the response body.
func (a *App) ReadBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	return string(body)
}

// ParseJSON unmarshals the response body into v.
func (a *App) ParseJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
}
