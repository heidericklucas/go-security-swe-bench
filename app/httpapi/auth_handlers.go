package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/heidericklucas/go-security-swe-bench/app/store"
)

func (s *server) register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string
		Password string
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	uid, err := s.store.CreateUser(store.User{Username: req.Username, PasswordHash: req.Password})
	if err != nil {
		http.Error(w, "conflict", http.StatusConflict)
		return
	}
	u, _ := s.store.GetUserByID(uid)
	writeJSON(w, http.StatusCreated, u)
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string
		Password string
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	u, err := s.store.GetUserByUsername(req.Username)
	if err != nil || u.PasswordHash != req.Password {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	token, _ := s.verifier.Issue(u.ID, u.IsAdmin, 24*time.Hour)
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (s *server) getMe(w http.ResponseWriter, r *http.Request) {
	u, _ := s.store.GetUserByID(claimsFrom(r).UserID)
	writeJSON(w, http.StatusOK, u)
}
