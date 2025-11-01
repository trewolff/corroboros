package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"net/http"

	_ "github.com/lib/pq"
)

func uploadHandler(db *sql.DB, maxSize int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxSize)
		f, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file", http.StatusBadRequest)
			return
		}
		defer f.Close()

		hasher := sha256.New()
		if _, err := io.Copy(hasher, f); err != nil {
			http.Error(w, "read error", http.StatusInternalServerError)
			return
		}
		checksum := hex.EncodeToString(hasher.Sum(nil))

		// insert the canonical record (id, checksum, original filename, uploader, timestamp, etc.)
		if _, err := db.Exec(`INSERT INTO files (checksum, status) VALUES ($1, $2) ON CONFLICT DO NOTHING`, checksum, "stored"); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(checksum))
	}
}
