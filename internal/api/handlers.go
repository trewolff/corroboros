package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

func uploadHandler(db *sql.DB, maxSize int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxSize)
		f, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file", http.StatusBadRequest)
			return
		}
		defer f.Close()

		// Get original filename and content type from the file header
		origFilename := header.Filename
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		// Get file size
		fileSize := header.Size

		// Get user ID from auth context (you might want to customize this based on your auth system)
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "missing user ID", http.StatusUnauthorized)
			return
		}

		// Get submitter's IP address
		submitterIP := r.Header.Get("X-Real-IP")
		if submitterIP == "" {
			submitterIP = r.Header.Get("X-Forwarded-For")
			if submitterIP == "" {
				submitterIP = r.RemoteAddr
			}
		}

		// Get submitter's hostname
		submitterHostname := r.Host

		// Get source machine details from headers
		sourceIP := r.Header.Get("X-Source-IP")
		sourceHostname := r.Header.Get("X-Source-Hostname")
		originalStoragePath := r.Header.Get("X-Original-Path")

		// Get upload ID if this is part of a batch upload
		uploadID := r.Header.Get("X-Upload-ID")

		// Validate file size against max size
		if fileSize > maxSize {
			http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
			return
		}

		// Calculate checksum
		hasher := sha256.New()
		if _, err := io.Copy(hasher, f); err != nil {
			http.Error(w, "read error", http.StatusInternalServerError)
			return
		}
		checksum := hex.EncodeToString(hasher.Sum(nil))

		// Seek back to beginning of file for potential future operations
		if _, err := f.Seek(0, 0); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// Generate storage path based on checksum and date for better organization
		timestamp := time.Now()
		datePath := timestamp.Format("2006/01/02")
		storagePath := "/storage/" + datePath + "/" + checksum[:2] + "/" + checksum[2:4] + "/" + checksum

		sqlResult, err := db.Exec(`
			INSERT INTO files (
				checksum, status, timestamp, original_filename, user_id, size, content_type, 
				storage_path, submitter_ip, submitter_hostname, source_ip, source_hostname,
				original_storage_path, upload_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
			ON CONFLICT (checksum) DO NOTHING
		`, checksum, "stored", timestamp, origFilename, userID, fileSize, contentType,
			storagePath, submitterIP, submitterHostname, sourceIP, sourceHostname,
			originalStoragePath, uploadID)

		if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}

		rowsAffected, err := sqlResult.RowsAffected()
		if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}

		if rowsAffected == 0 {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(checksum))
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(checksum))
	}
}
