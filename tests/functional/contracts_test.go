package functional_test

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/heidericklucas/go-security-swe-bench/tests/harness"
)

// TestContract_OwnerCanReadOwnNote tests that bob can read his own private note.
func TestContract_OwnerCanReadOwnNote(t *testing.T) {
	a := harness.Start(t)

	// Bob should be able to read his own private note
	bobToken := a.Token(t, a.Data.BobID, false)
	noteIDStr := strconv.FormatInt(a.Data.BobPrivateNoteID, 10)
	resp := a.Do(t, http.MethodGet, "/notes/"+noteIDStr, bobToken, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := a.ReadBody(t, resp)
	if !strings.Contains(body, "BOB_SECRET_NOTE") {
		t.Fatalf("expected BOB_SECRET_NOTE in response, got: %s", body)
	}
}

// TestContract_SearchReturnsOwnerMatches tests that search returns only owner notes.
func TestContract_SearchReturnsOwnerMatches(t *testing.T) {
	a := harness.Start(t)

	// Alice searches for groceries - should find her note
	aliceToken := a.Token(t, a.Data.AliceID, false)
	resp := a.Do(t, http.MethodGet, "/notes/search?q=groceries", aliceToken, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var notes []map[string]any
	a.ParseJSON(t, resp, &notes)

	if len(notes) != 1 {
		t.Fatalf("expected 1 note, got %d", len(notes))
	}

	if notes[0]["Title"] != "groceries" {
		t.Fatalf("expected groceries, got %v", notes[0]["Title"])
	}
}

// TestContract_ValidTokenAccepted tests that a valid RS256 token is accepted.
func TestContract_ValidTokenAccepted(t *testing.T) {
	a := harness.Start(t)

	// Generate a valid token for alice
	aliceToken := a.Token(t, a.Data.AliceID, false)

	// Call /me with the token
	resp := a.Do(t, http.MethodGet, "/me", aliceToken, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var user map[string]any
	a.ParseJSON(t, resp, &user)

	if user["Username"] != "alice" {
		t.Fatalf("expected alice, got %v", user["Username"])
	}
}

// TestContract_ExportRendersBenignNote tests that note export renders the title.
func TestContract_ExportRendersBenignNote(t *testing.T) {
	a := harness.Start(t)

	// Alice exports her note (note 1)
	aliceToken := a.Token(t, a.Data.AliceID, false)
	resp := a.Do(t, http.MethodGet, "/notes/1/export", aliceToken, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := a.ReadBody(t, resp)
	if !strings.Contains(body, "groceries") {
		t.Fatalf("expected groceries in export, got: %s", body)
	}
}

// TestContract_BenignFileDownload tests that authorized users can download files.
func TestContract_BenignFileDownload(t *testing.T) {
	a := harness.Start(t)

	// Alice downloads avatar.png
	aliceToken := a.Token(t, a.Data.AliceID, false)
	resp := a.Do(t, http.MethodGet, "/files?name=avatar.png", aliceToken, nil)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body := a.ReadBody(t, resp)
	if !strings.Contains(body, "PNGDATA") {
		t.Fatalf("expected PNGDATA in response, got: %s", body)
	}
}
