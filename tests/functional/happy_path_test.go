package functional_test

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/heidericklucas/go-security-swe-bench/tests/harness"
)

// TestHappyPath_EndToEnd tests the full user flow.
func TestHappyPath_EndToEnd(t *testing.T) {
	a := harness.Start(t)

	// Register a new user
	registerResp := a.Do(t, http.MethodPost, "/register", "", map[string]string{
		"Username": "testuser",
		"Password": "testpass",
	})
	defer registerResp.Body.Close()

	if registerResp.StatusCode != http.StatusCreated {
		t.Fatalf("register failed: expected 201, got %d", registerResp.StatusCode)
	}

	var userResp map[string]any
	a.ParseJSON(t, registerResp, &userResp)
	if _, ok := userResp["ID"].(float64); !ok {
		t.Fatalf("register response missing numeric ID: %v", userResp)
	}

	// Login
	loginResp := a.Do(t, http.MethodPost, "/login", "", map[string]string{
		"Username": "testuser",
		"Password": "testpass",
	})
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: expected 200, got %d", loginResp.StatusCode)
	}

	var loginBody map[string]string
	a.ParseJSON(t, loginResp, &loginBody)
	token := loginBody["token"]

	// Create a note
	createResp := a.Do(t, http.MethodPost, "/notes", token, map[string]string{
		"Title": "My Test Note",
		"Body":  "This is a test note",
	})
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create note failed: expected 201, got %d", createResp.StatusCode)
	}

	var noteResp map[string]any
	a.ParseJSON(t, createResp, &noteResp)
	noteID := int64(noteResp["ID"].(float64))

	// Read the note
	noteIDStr := strconv.FormatInt(noteID, 10)
	getResp := a.Do(t, http.MethodGet, "/notes/"+noteIDStr, token, nil)
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get note failed: expected 200, got %d", getResp.StatusCode)
	}

	var getNoteBody map[string]any
	a.ParseJSON(t, getResp, &getNoteBody)

	if getNoteBody["Title"] != "My Test Note" {
		t.Fatalf("expected title My Test Note, got %v", getNoteBody["Title"])
	}

	// Export the note
	exportResp := a.Do(t, http.MethodGet, "/notes/"+noteIDStr+"/export", token, nil)
	defer exportResp.Body.Close()

	if exportResp.StatusCode != http.StatusOK {
		t.Fatalf("export note failed: expected 200, got %d", exportResp.StatusCode)
	}

	exportBody := a.ReadBody(t, exportResp)
	if !strings.Contains(exportBody, "My Test Note") {
		t.Fatalf("expected My Test Note in export, got: %s", exportBody)
	}

	// Download a file
	fileResp := a.Do(t, http.MethodGet, "/files?name=avatar.png", token, nil)
	defer fileResp.Body.Close()

	if fileResp.StatusCode != http.StatusOK {
		t.Fatalf("get file failed: expected 200, got %d", fileResp.StatusCode)
	}

	fileBody := a.ReadBody(t, fileResp)
	if !strings.Contains(fileBody, "PNGDATA") {
		t.Fatalf("expected PNGDATA in file, got: %s", fileBody)
	}
}
