package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/trewolff/corroboros/internal/database"
)

type UploadHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type HandlerDependencies struct {
	DB            DB
	Logger        *slog.Logger
	MaxIntakeSize int64
}

// DB is the minimal database interface used by the handlers so tests and
// other implementations can provide a mock.
type DB interface {
	CreateRecords(record database.Record) (sql.Result, error)
	GetRecords() ([]database.Record, error)
	GetRecordsByUserID(userID string) ([]database.Record, error)
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
func (h *HandlerDependencies) uploadCoreLogic(input UploadInput) UploadResult {
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
	record := database.Record{
		Checksum:            checksum,
		Status:              "stored",
		Timestamp:           timestamp,
		OriginalFilename:    input.OrigFilename,
		UserID:              input.UserID,
		FileSize:            input.FileSize,
		ContentType:         input.ContentType,
		StoragePath:         storagePath,
		SubmitterIP:         input.SubmitterIP,
		SubmitterHostname:   input.SubmitterHostname,
		SourceIP:            input.SourceIP,
		SourceHostname:      input.SourceHostname,
		OriginalStoragePath: input.OriginalStoragePath,
		UploadID:            input.UploadID,
	}
	slog.Info("storing record", "record", record)
	//records := []database.Record{record}
	sqlResult, err := h.DB.CreateRecords(record)
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
		result := h.uploadCoreLogic(input)
		if result.Err != nil {
			http.Error(w, result.Status, result.Code)
			return
		}
		w.WriteHeader(result.Code)
		w.Write([]byte(result.Checksum))
	}
}

// getRecordsHandler handles the /records endpoint
func (h *HandlerDependencies) getRecordsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: implement pagination, filtering, etc.
		// Implement authentication/authorization as needed
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "expected_api_key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		userID, err := getUserIDFromAPIKey(apiKey)
		if err != nil || userID.String() == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		records, err := h.DB.GetRecordsByUserID(userID.String())
		if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}
		// Serialize records to JSON and write to response
		err = json.NewEncoder(w).Encode(records)
		if err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			return
		}
	}
}
