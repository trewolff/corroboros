package api

import (
	"bytes"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/trewolff/corroboros/internal/database"
)

type mockDB struct {
	createFunc      func(rec database.Record) (sql.Result, error)
	getFunc         func() ([]database.Record, error)
	getByUserIDFunc func(userID string) ([]database.Record, error)
}

// CreateRecords mocks the CreateRecords method
func (m *mockDB) CreateRecords(rec database.Record) (sql.Result, error) {
	if m.createFunc != nil {
		return m.createFunc(rec)
	}
	return nil, nil
}

// GetRecords mocks the GetRecords method
func (m *mockDB) GetRecords() ([]database.Record, error) {
	if m.getFunc != nil {
		return m.getFunc()
	}
	return nil, nil
}

func (m *mockDB) GetRecordsByUserID(userID string) ([]database.Record, error) {
	if m.getFunc != nil {
		return m.getFunc()
	}
	return nil, nil
}

type mockResult struct {
	affected int64
	err      error
}

func (m *mockResult) LastInsertId() (int64, error) { return 0, nil }
func (m *mockResult) RowsAffected() (int64, error) { return m.affected, m.err }

func TestUploadCoreLogic(t *testing.T) {
	// Helper to create UploadInput with a given file content and size
	makeInput := func(content string, size int64) UploadInput {
		return UploadInput{
			File:                bytes.NewReader([]byte(content)),
			OrigFilename:        "test.txt",
			ContentType:         "text/plain",
			FileSize:            size,
			UserID:              "user1",
			SubmitterIP:         "1.2.3.4",
			SubmitterHostname:   "submitterhost",
			SourceIP:            "5.6.7.8",
			SourceHostname:      "sourcehost",
			OriginalStoragePath: "/original/path",
			UploadID:            "upload123",
			MaxIntakeSize:       100,
		}
	}

	t.Run("file too large", func(t *testing.T) {
		input := makeInput("abc", 200)
		hd := &HandlerDependencies{DB: &mockDB{}, Logger: nil, MaxIntakeSize: 100}
		res := hd.uploadCoreLogic(input)
		if res.Code != 413 || !strings.Contains(res.Status, "file too large") {
			t.Errorf("expected file too large error, got %+v", res)
		}
	})

	t.Run("read error", func(t *testing.T) {
		badReader := &errReader{}
		input := makeInput("", 10)
		input.File = badReader
		hd := &HandlerDependencies{DB: &mockDB{}, Logger: nil, MaxIntakeSize: 100}
		res := hd.uploadCoreLogic(input)
		if res.Code != 500 || !strings.Contains(res.Status, "read error") {
			t.Errorf("expected read error, got %+v", res)
		}
	})

	t.Run("seek error", func(t *testing.T) {
		input := makeInput("abc", 3)
		input.File = &seekErrReader{bytes.NewReader([]byte("abc"))}
		hd := &HandlerDependencies{DB: &mockDB{}, Logger: nil, MaxIntakeSize: 100}
		res := hd.uploadCoreLogic(input)
		if res.Code != 500 || !strings.Contains(res.Status, "internal error") {
			t.Errorf("expected seek error, got %+v", res)
		}
	})

	t.Run("database error", func(t *testing.T) {
		input := makeInput("abc", 3)
		mdb := &mockDB{
			createFunc: func(rec database.Record) (sql.Result, error) {
				return nil, errors.New("db fail")
			},
		}
		hd := &HandlerDependencies{DB: mdb, Logger: nil, MaxIntakeSize: 100}
		res := hd.uploadCoreLogic(input)
		if res.Code != 500 || !strings.Contains(res.Status, "database error") {
			t.Errorf("expected db error, got %+v", res)
		}
	})

	t.Run("rows affected error", func(t *testing.T) {
		input := makeInput("abc", 3)
		mdb := &mockDB{
			createFunc: func(rec database.Record) (sql.Result, error) {
				return &mockResult{affected: 0, err: errors.New("rows fail")}, nil
			},
		}
		hd := &HandlerDependencies{DB: mdb, Logger: nil, MaxIntakeSize: 100}
		res := hd.uploadCoreLogic(input)
		if res.Code != 500 || !strings.Contains(res.Status, "database error") {
			t.Errorf("expected rows error, got %+v", res)
		}
	})

	t.Run("new file created", func(t *testing.T) {
		input := makeInput("abc", 3)
		mdb := &mockDB{
			createFunc: func(rec database.Record) (sql.Result, error) {
				return &mockResult{affected: 1, err: nil}, nil
			},
		}
		hd := &HandlerDependencies{DB: mdb, Logger: nil, MaxIntakeSize: 100}
		res := hd.uploadCoreLogic(input)
		if res.Code != 201 || res.Status != "created" || res.Checksum == "" {
			t.Errorf("expected created, got %+v", res)
		}
	})

	t.Run("file already exists", func(t *testing.T) {
		input := makeInput("abc", 3)
		mdb := &mockDB{
			createFunc: func(rec database.Record) (sql.Result, error) {
				return &mockResult{affected: 0, err: nil}, nil
			},
		}
		hd := &HandlerDependencies{DB: mdb, Logger: nil, MaxIntakeSize: 100}
		res := hd.uploadCoreLogic(input)
		if res.Code != 200 || res.Status != "exists" || res.Checksum == "" {
			t.Errorf("expected exists, got %+v", res)
		}
	})
}

type errReader struct{}

func (e *errReader) Read(p []byte) (int, error)                   { return 0, errors.New("read fail") }
func (e *errReader) Seek(offset int64, whence int) (int64, error) { return 0, nil }

// seekErrReader simulates a file that errors on Seek
// wraps a bytes.Reader
type seekErrReader struct{ *bytes.Reader }

func (s *seekErrReader) Read(p []byte) (int, error) { return s.Reader.Read(p) }
func (s *seekErrReader) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("seek fail")
}
