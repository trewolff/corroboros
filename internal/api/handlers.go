package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type UploadHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type DB interface {
	Exec(query string, args ...any) (sql.Result, error)
}

type HandlerDependencies struct {
	DB            DB
	Logger        *slog.Logger
	MaxIntakeSize int64
}

// UploadResult holds the result of a file upload operation
type UploadResult struct {
	Checksum string
	Status   string
	Code     int
	Err      error
}

// UploadInput holds all extracted upload parameters
type UploadInput struct {
	File                io.ReadSeeker
	OrigFilename        string
	ContentType         string
	FileSize            int64
	UserID              string
	SubmitterIP         string
	SubmitterHostname   string
	SourceIP            string
	SourceHostname      string
	OriginalStoragePath string
	UploadID            string
	MaxIntakeSize       int64
}

// uploadCoreLogic performs the core upload logic, returns UploadResult
func uploadCoreLogic(db DB, input UploadInput) UploadResult {
	if input.FileSize > input.MaxIntakeSize {
		return UploadResult{"", "file too large", http.StatusRequestEntityTooLarge, nil}
	}
	hasher := sha256.New()
	if _, err := io.Copy(hasher, input.File); err != nil {
		return UploadResult{"", "read error", http.StatusInternalServerError, err}
	}
	checksum := hex.EncodeToString(hasher.Sum(nil))
	if _, err := input.File.Seek(0, 0); err != nil {
		return UploadResult{"", "internal error", http.StatusInternalServerError, err}
	}
	timestamp := time.Now()
	datePath := timestamp.Format("2006/01/02")
	storagePath := "/storage/" + datePath + "/" + checksum[:2] + "/" + checksum[2:4] + "/" + checksum
	fmt.Println("input", input)
	sqlResult, err := db.Exec(`
	       INSERT INTO files (
		       checksum, status, timestamp, original_filename, user_id, size, content_type, 
		       storage_path, submitter_ip, submitter_hostname, source_ip, source_hostname,
		       original_storage_path, upload_id
	       ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	       ON CONFLICT (checksum) DO NOTHING
       `, checksum, "stored", timestamp, input.OrigFilename, input.UserID, input.FileSize, input.ContentType,
		storagePath, input.SubmitterIP, input.SubmitterHostname, input.SourceIP, input.SourceHostname,
		input.OriginalStoragePath, input.UploadID)
	if err != nil {
		return UploadResult{"", "database error", http.StatusInternalServerError, err}
	}
	rowsAffected, err := sqlResult.RowsAffected()
	if err != nil {
		return UploadResult{"", "database error", http.StatusInternalServerError, err}
	}
	code := http.StatusCreated
	status := "created"
	if rowsAffected == 0 {
		code = http.StatusOK
		status = "exists"
	}
	return UploadResult{checksum, status, code, nil}
}

// uploadHandler is the HTTP handler, now thin and testable
func (h *HandlerDependencies) uploadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, h.MaxIntakeSize)
		f, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file", http.StatusBadRequest)
			return
		}
		defer f.Close()

		origFilename := header.Filename
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		fileSize := header.Size
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			http.Error(w, "missing user ID", http.StatusUnauthorized)
			return
		}
		submitterIP := r.Header.Get("X-Real-IP")
		if submitterIP == "" {
			submitterIP = r.Header.Get("X-Forwarded-For")
			if submitterIP == "" {
				submitterIP = r.RemoteAddr
			}
		}
		submitterHostname := r.Host
		sourceIP := r.Header.Get("X-Source-IP")
		sourceHostname := r.Header.Get("X-Source-Hostname")
		originalStoragePath := r.Header.Get("X-Original-Path")
		uploadID := r.Header.Get("X-Upload-ID")

		input := UploadInput{
			File:                f,
			OrigFilename:        origFilename,
			ContentType:         contentType,
			FileSize:            fileSize,
			UserID:              userID,
			SubmitterIP:         submitterIP,
			SubmitterHostname:   submitterHostname,
			SourceIP:            sourceIP,
			SourceHostname:      sourceHostname,
			OriginalStoragePath: originalStoragePath,
			UploadID:            uploadID,
			MaxIntakeSize:       h.MaxIntakeSize,
		}
		result := uploadCoreLogic(h.DB, input)
		if result.Err != nil {
			http.Error(w, result.Status, result.Code)
			return
		}
		w.WriteHeader(result.Code)
		w.Write([]byte(result.Checksum))
	}
}
