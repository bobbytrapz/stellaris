package localisation

import (
	"database/sql"
	"fmt"
	"strings"
	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

func NewStore() (*Store, error) {
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE strings (
			key TEXT PRIMARY KEY,
			base_key TEXT,
			variant TEXT,
			suffix TEXT,
			version TEXT,
			category TEXT,
			text TEXT,
			filename TEXT,
			modified BOOLEAN DEFAULT 0
		);
		CREATE VIRTUAL TABLE strings_fts USING fts5(
			key,
			base_key,
			text,
			content='strings',
			content_rowid='rowid'
		);
		
		CREATE TRIGGER strings_ai AFTER INSERT ON strings BEGIN
			INSERT INTO strings_fts(rowid, key, base_key, text) VALUES (new.rowid, new.key, new.base_key, new.text);
		END;
		CREATE TRIGGER strings_ad AFTER DELETE ON strings BEGIN
			INSERT INTO strings_fts(strings_fts, rowid, key, base_key, text) VALUES ('delete', old.rowid, old.key, old.base_key, old.text);
		END;
		CREATE TRIGGER strings_au AFTER UPDATE ON strings BEGIN
			INSERT INTO strings_fts(strings_fts, rowid, key, base_key, text) VALUES ('delete', old.rowid, old.key, old.base_key, old.text);
			INSERT INTO strings_fts(rowid, key, base_key, text) VALUES (new.rowid, new.key, new.base_key, new.text);
		END;
	`)
	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) InsertStrings(strings []LocString) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO strings (key, base_key, variant, suffix, version, category, text, filename, modified)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, str := range strings {
		_, err := stmt.Exec(str.Key, str.BaseKey, str.Variant, str.Suffix, str.Version, str.Category, str.Text, str.Filename, str.Modified)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) Search(query string) ([]LocString, error) {
	if query == "" {
		rows, err := s.db.Query(`
			SELECT key, base_key, variant, suffix, version, category, text, filename, modified
			FROM strings
			LIMIT 100
		`)
		if err != nil {
			return nil, err
		}
		return scanRows(rows)
	}

	// Fetch all variants/suffixes for any base_key that matched the search query, ordered by BM25 rank
	rankRows, err := s.db.Query(`
		SELECT base_key
		FROM strings_fts
		WHERE strings_fts MATCH ?
		ORDER BY bm25(strings_fts, 10.0, 10.0, 1.0)
		LIMIT 200
	`, query)
	if err == nil {
		defer rankRows.Close()
		var baseKeys []string
		seen := make(map[string]bool)
		for rankRows.Next() {
			var bk string
			if err := rankRows.Scan(&bk); err == nil {
				if !seen[bk] {
					seen[bk] = true
					baseKeys = append(baseKeys, bk)
					if len(baseKeys) >= 50 {
						break
					}
				}
			}
		}

		if len(baseKeys) > 0 {
			placeholders := make([]string, len(baseKeys))
			args := make([]interface{}, len(baseKeys))
			for i, bk := range baseKeys {
				placeholders[i] = "?"
				args[i] = bk
			}
			
			rows, err := s.db.Query(fmt.Sprintf(`
				SELECT key, base_key, variant, suffix, version, category, text, filename, modified
				FROM strings
				WHERE base_key IN (%s)
			`, strings.Join(placeholders, ",")), args...)
			
			if err == nil {
				strs, scanErr := scanRows(rows)
				if scanErr == nil {
					// We need to preserve the ranked order since IN doesn't preserve order
					var orderedStrs []LocString
					for _, bk := range baseKeys {
						for _, s := range strs {
							if s.BaseKey == bk {
								orderedStrs = append(orderedStrs, s)
							}
						}
					}
					return orderedStrs, nil
				}
			}
		}
	}

	// Fallback if FTS failed or returned 0 results
	likeQuery := "%" + query + "%"
	fallbackRows, err := s.db.Query(`
			SELECT key, base_key, variant, suffix, version, category, text, filename, modified
			FROM strings
			WHERE base_key IN (
				SELECT DISTINCT base_key
				FROM strings
				WHERE key LIKE ? OR text LIKE ?
				LIMIT 50
			)
		`, likeQuery, likeQuery)
	if err != nil {
		return nil, err
	}
	return scanRows(fallbackRows)
}

func (s *Store) GetCategories() ([]string, error) {
	rows, err := s.db.Query(`SELECT DISTINCT category FROM strings ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err == nil {
			cats = append(cats, c)
		}
	}
	return cats, nil
}

func (s *Store) GetByCategory(category string) ([]LocString, error) {
	// Let's limit to 100 distinct base keys to avoid overloading the UI
	rows, err := s.db.Query(`
		SELECT key, base_key, variant, suffix, version, category, text, filename, modified
		FROM strings
		WHERE base_key IN (
			SELECT DISTINCT base_key
			FROM strings
			WHERE category = ?
			LIMIT 50
		)
	`, category)
	if err != nil {
		return nil, err
	}
	return scanRows(rows)
}

func (s *Store) Update(key string, text string) error {
	_, err := s.db.Exec(`
		UPDATE strings
		SET text = ?, modified = 1
		WHERE key = ?
	`, text, key)
	return err
}

func (s *Store) GetModified() ([]LocString, error) {
	rows, err := s.db.Query(`
		SELECT key, base_key, variant, suffix, version, category, text, filename, modified
		FROM strings
		WHERE modified = 1
	`)
	if err != nil {
		return nil, err
	}
	return scanRows(rows)
}

func scanRows(rows *sql.Rows) ([]LocString, error) {
	defer rows.Close()
	var results []LocString
	for rows.Next() {
		var l LocString
		if err := rows.Scan(&l.Key, &l.BaseKey, &l.Variant, &l.Suffix, &l.Version, &l.Category, &l.Text, &l.Filename, &l.Modified); err != nil {
			return nil, err
		}
		results = append(results, l)
	}
	return results, nil
}
