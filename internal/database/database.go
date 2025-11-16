package database

import (
	"database/sql"
	"time"
)

type Record struct {
	ID                  int       `json:"id"`
	Checksum            string    `json:"checksum"`
	Status              string    `json:"status"`
	Timestamp           time.Time `json:"timestamp"`
	OriginalFilename    string    `json:"original_filename"`
	UserID              string    `json:"user_id"`
	FileSize            int64     `json:"file_size"`
	ContentType         string    `json:"content_type"`
	StoragePath         string    `json:"storage_path"`
	SubmitterIP         string    `json:"submitter_ip"`
	SubmitterHostname   string    `json:"submitter_hostname"`
	SourceIP            string    `json:"source_ip"`
	SourceHostname      string    `json:"source_hostname"`
	OriginalStoragePath string    `json:"original_storage_path"`
	UploadID            string    `json:"upload_id"`
}

type Database struct {
	Connection *sql.DB
}

// NewDatabase initializes a new Database instance
func NewDatabase(connStr string) (*Database, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &Database{Connection: db}, nil
}

// GetRecords retrieves all records from the database
func (d *Database) GetRecords() ([]Record, error) {
	rows, err := d.Connection.Query("SELECT * FROM records")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		if err := rows.Scan(
			&record.ID,
			&record.Checksum,
			&record.Status,
			&record.Timestamp,
			&record.OriginalFilename,
			&record.UserID,
			&record.FileSize,
			&record.ContentType,
			&record.StoragePath,
			&record.SubmitterIP,
			&record.SubmitterHostname,
			&record.SourceIP,
			&record.SourceHostname,
			&record.OriginalStoragePath,
			&record.UploadID,
		); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

// CreateRecords inserts a new record into the database
func (d *Database) CreateRecords(record Record) (sql.Result, error) {
	sqlResult, err := d.Connection.Exec(`
	       INSERT INTO files (
		       checksum, status, timestamp, original_filename, user_id, size, content_type, 
		       storage_path, submitter_ip, submitter_hostname, source_ip, source_hostname,
		       original_storage_path, upload_id
	       ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	       ON CONFLICT (checksum) DO NOTHING
       `, record.Checksum, record.Status, record.Timestamp, record.OriginalFilename,
		record.UserID, record.FileSize, record.ContentType, record.StoragePath,
		record.SubmitterIP, record.SubmitterHostname, record.SourceIP, record.SourceHostname,
		record.OriginalStoragePath, record.UploadID)
	return sqlResult, err
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.Connection.Close()
}

func (d *Database) GetRecordsByUserID(userID string) ([]Record, error) {
	rows, err := d.Connection.Query("SELECT * FROM records WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		if err := rows.Scan(
			&record.ID,
			&record.Checksum,
			&record.Status,
			&record.Timestamp,
			&record.OriginalFilename,
			&record.UserID,
			&record.FileSize,
			&record.ContentType,
			&record.StoragePath,
			&record.SubmitterIP,
			&record.SubmitterHostname,
			&record.SourceIP,
			&record.SourceHostname,
			&record.OriginalStoragePath,
			&record.UploadID,
		); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}
