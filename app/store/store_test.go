package store_test

import (
	"testing"
	"time"

	"github.com/heidericklucas/go-security-swe-bench/app/clock"
	"github.com/heidericklucas/go-security-swe-bench/app/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.New(":memory:", clock.Fixed(time.Unix(1_700_000_000, 0)))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestSearchNotes_ReturnsOnlyOwnerMatches(t *testing.T) {
	s := newTestStore(t)
	alice, _ := s.CreateUser(store.User{Username: "alice", PasswordHash: "x"})
	bob, _ := s.CreateUser(store.User{Username: "bob", PasswordHash: "x"})
	_, _ = s.CreateNote(store.Note{OwnerID: alice, Title: "groceries", Body: "milk", Private: true})
	_, _ = s.CreateNote(store.Note{OwnerID: bob, Title: "secret", Body: "bob-only", Private: true})

	got, err := s.SearchNotes(alice, "groceries")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(got) != 1 || got[0].Title != "groceries" {
		t.Fatalf("want alice's single note, got %+v", got)
	}
}
