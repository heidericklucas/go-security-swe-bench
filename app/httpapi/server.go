package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/heidericklucas/go-security-swe-bench/app/auth"
	"github.com/heidericklucas/go-security-swe-bench/app/fetch"
	"github.com/heidericklucas/go-security-swe-bench/app/store"
)

type server struct {
	store    *store.Store
	verifier *auth.Verifier
	fetcher  *fetch.Fetcher
	filesDir string
}

type ctxKey int

const claimsKey ctxKey = 0

func New(s *store.Store, v *auth.Verifier, f *fetch.Fetcher, filesDir string) http.Handler {
	srv := &server{store: s, verifier: v, fetcher: f, filesDir: filesDir}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", srv.register)
	mux.HandleFunc("POST /login", srv.login)
	mux.Handle("GET /me", srv.requireAuth(srv.getMe))
	mux.Handle("GET /notes", srv.requireAuth(srv.listNotes))
	mux.Handle("POST /notes", srv.requireAuth(srv.createNote))
	mux.Handle("GET /notes/{id}", srv.requireAuth(srv.getNote))
	mux.Handle("DELETE /notes/{id}", srv.requireAuth(srv.deleteNote))
	mux.Handle("GET /notes/search", srv.requireAuth(srv.searchNotes))
	mux.Handle("POST /notes/import", srv.requireAuth(srv.importNote))
	mux.Handle("GET /notes/{id}/export", srv.requireAuth(srv.exportNote))
	mux.Handle("GET /files", srv.requireAuth(srv.getFile))
	return mux
}

func (s *server) requireAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		tok := strings.TrimPrefix(h, "Bearer ")
		if tok == h || tok == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		claims, err := s.verifier.Verify(tok)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), claimsKey, claims)))
	})
}

func claimsFrom(r *http.Request) *auth.Claims {
	c, _ := r.Context().Value(claimsKey).(*auth.Claims)
	return c
}
