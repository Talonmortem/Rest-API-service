package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/Talonmortem/Rest-API-service/internal/lib/random"
	"github.com/Talonmortem/Rest-API-service/internal/storage"
	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"
	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	stmt, err := db.Prepare(
		`CREATE TABLE IF NOT EXISTS url(id INTEGER PRIMARY KEY, alias TEXT NOT NULL UNIQUE, url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("SELECT count(alias) FROM url where alias = ?")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	countaliace, err := stmt.Exec(alias)
	var count int

	err = stmt.QueryRow(alias).Scan(&count)
	for count != 0 {
		alias = random.NewRandomString(len(alias))
		stmt, err := s.db.Prepare("SELECT count(alias) FROM url where alias = ?")
		if err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
		countaliace, err = stmt.Exec(alias)
		count, err = fmt.Scan(&countaliace)
	}
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err = s.db.Prepare("INSERT INTO url(alias, url) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(alias, urlToSave)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s:failed to get last insert id: %w", op, err)
	}
	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"
	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resURL string
	err = stmt.QueryRow(alias).Scan(&resURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
	}
	if err != nil {
		return "", fmt.Errorf("%s: execute statement %w", op, err)
	}
	return resURL, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.sqlite.DeleteURL"
	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = stmt.Exec(alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

/*func (s *Storage) CheckAlias(alias string) error {
	const op = "storage.sqlite.CheckAlias"
	stmt, err := s.db.Prepare("SELECT count(alias) FROM url where alias = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	countaliace, err := stmt.Exec(alias)
	count, err := fmt.Scan(&countaliace)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if count > 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrAliasExist)
	}
	return nil
}

*/
