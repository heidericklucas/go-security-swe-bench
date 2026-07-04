package store

import (
	"database/sql"
	"fmt"

	"github.com/heidericklucas/go-security-swe-bench/app/clock"
	_ "modernc.org/sqlite"
)

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	IsAdmin      bool
}

type Note struct {
	ID      int64
	OwnerID int64
	Title   string
	Body    string
	Private bool
}

type Store struct {
	db  *sql.DB
	clk clock.Clock
}

func New(dsn string, clk clock.Clock) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // sqlite: serialize; also makes :memory: stable
	const ddl = `
CREATE TABLE users (id INTEGER PRIMARY KEY, username TEXT UNIQUE, password_hash TEXT, is_admin INTEGER DEFAULT 0);
CREATE TABLE notes (id INTEGER PRIMARY KEY, owner_id INTEGER, title TEXT, body TEXT, private INTEGER DEFAULT 0);`
	if _, err := db.Exec(ddl); err != nil {
		return nil, err
	}
	return &Store{db: db, clk: clk}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) CreateUser(u User) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO users(username,password_hash,is_admin) VALUES(?,?,?)`,
		u.Username, u.PasswordHash, u.IsAdmin)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) GetUserByUsername(username string) (*User, error) {
	row := s.db.QueryRow(`SELECT id,username,password_hash,is_admin FROM users WHERE username=?`, username)
	return scanUser(row)
}

func (s *Store) GetUserByID(id int64) (*User, error) {
	row := s.db.QueryRow(`SELECT id,username,password_hash,is_admin FROM users WHERE id=?`, id)
	return scanUser(row)
}

func (s *Store) SetAdmin(id int64, v bool) error {
	_, err := s.db.Exec(`UPDATE users SET is_admin=? WHERE id=?`, v, id)
	return err
}

func (s *Store) CreateNote(n Note) (int64, error) {
	res, err := s.db.Exec(`INSERT INTO notes(owner_id,title,body,private) VALUES(?,?,?,?)`,
		n.OwnerID, n.Title, n.Body, n.Private)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) GetNoteByID(id int64) (*Note, error) {
	row := s.db.QueryRow(`SELECT id,owner_id,title,body,private FROM notes WHERE id=?`, id)
	return scanNote(row)
}

func (s *Store) ListNotesByOwner(ownerID int64) ([]Note, error) {
	rows, err := s.db.Query(`SELECT id,owner_id,title,body,private FROM notes WHERE owner_id=? ORDER BY id`, ownerID)
	if err != nil {
		return nil, err
	}
	return scanNotes(rows)
}

// SearchNotes returns the caller's notes whose title contains q.
func (s *Store) SearchNotes(ownerID int64, q string) ([]Note, error) {
	query := fmt.Sprintf(
		`SELECT id,owner_id,title,body,private FROM notes WHERE owner_id=%d AND title LIKE '%%%s%%' ORDER BY id`,
		ownerID, q)
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	return scanNotes(rows)
}

func (s *Store) DeleteNote(id int64) error {
	_, err := s.db.Exec(`DELETE FROM notes WHERE id=?`, id)
	return err
}

func scanUser(row *sql.Row) (*User, error) {
	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.IsAdmin); err != nil {
		return nil, err
	}
	return &u, nil
}

func scanNote(row *sql.Row) (*Note, error) {
	var n Note
	if err := row.Scan(&n.ID, &n.OwnerID, &n.Title, &n.Body, &n.Private); err != nil {
		return nil, err
	}
	return &n, nil
}

func scanNotes(rows *sql.Rows) ([]Note, error) {
	defer rows.Close()
	var out []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.OwnerID, &n.Title, &n.Body, &n.Private); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
