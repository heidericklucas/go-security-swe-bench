package functional_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"github.com/heidericklucas/go-security-swe-bench/tests/harness"
)

// PASS_TO_PASS: a legitimate allowlisted external import returns content. GREEN
// on base AND golden (external.test resolves to a public IP the golden permits).
func TestContract_ImportAllowlistedReturnsContent(t *testing.T) {
	a := harness.Start(t)
	external := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("EXTERNAL_OK"))
	}))
	defer external.Close()

	backend := strings.TrimPrefix(external.URL, "http://")
	a.Fetcher.Resolver = func(_ context.Context, host string) ([]netip.Addr, error) {
		if host == "external.test" {
			return []netip.Addr{netip.MustParseAddr("93.184.216.34")}, nil
		}
		return nil, &net.DNSError{Err: "no such host", Name: host}
	}
	a.Fetcher.Dial = func(ctx context.Context, network string, ip netip.Addr, port string) (net.Conn, error) {
		var d net.Dialer
		return d.DialContext(ctx, network, backend)
	}

	tok := a.Token(t, a.Data.AliceID, false)
	resp := a.Do(t, "POST", "/notes/import?url=http://external.test/", tok, nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("allowlisted import: want 200, got %d", resp.StatusCode)
	}
	var out map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if !strings.Contains(out["content"], "EXTERNAL_OK") {
		t.Fatalf("allowlisted import should return content, got %+v", out)
	}
}
