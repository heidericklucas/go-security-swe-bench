package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/heidericklucas/go-security-swe-bench/app/auth"
	"github.com/heidericklucas/go-security-swe-bench/app/clock"
)

func testVerifier(t *testing.T) *auth.Verifier {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return auth.NewVerifier(auth.Keys{RSAPrivate: priv, RSAPublic: &priv.PublicKey},
		clock.Fixed(time.Unix(1_700_000_000, 0)))
}

func TestIssueThenVerify_RoundTrips(t *testing.T) {
	v := testVerifier(t)
	tok, err := v.Issue(42, false, time.Hour)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	claims, err := v.Verify(tok)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.UserID != 42 || claims.IsAdmin {
		t.Fatalf("bad claims: %+v", claims)
	}
}
