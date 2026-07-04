package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"text/template"

	"github.com/heidericklucas/go-security-swe-bench/app/store"
)

func (s *server) listNotes(w http.ResponseWriter, r *http.Request) {
	notes, _ := s.store.ListNotesByOwner(claimsFrom(r).UserID)
	writeJSON(w, http.StatusOK, notes)
}

func (s *server) createNote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string
		Body  string
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	noteID, _ := s.store.CreateNote(store.Note{
		OwnerID: claimsFrom(r).UserID,
		Title:   req.Title,
		Body:    req.Body,
	})
	n, _ := s.store.GetNoteByID(noteID)
	writeJSON(w, http.StatusCreated, n)
}

func (s *server) getNote(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	n, err := s.store.GetNoteByID(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (s *server) deleteNote(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	n, err := s.store.GetNoteByID(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if claimsFrom(r).UserID != n.OwnerID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	_ = s.store.DeleteNote(id)
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) searchNotes(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	notes, err := s.store.SearchNotes(claimsFrom(r).UserID, q)
	if err != nil {
		http.Error(w, "search failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, notes)
}

func (s *server) importNote(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.Query().Get("url")
	body, err := s.fetcher.Get(r.Context(), raw)
	if err != nil {
		http.Error(w, "import failed", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"content": string(body)})
}

func (s *server) exportNote(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	n, err := s.store.GetNoteByID(id)
	if err != nil || claimsFrom(r).UserID != n.OwnerID {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	const tpl = `<!doctype html><h1>{{.Title}}</h1><div>{{.Body}}</div>`
	t := template.Must(template.New("export").Parse(tpl))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = t.Execute(w, n)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
