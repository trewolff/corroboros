package database

import (
	"database/sql"

	"github.com/trewolff/corroboros/internal/config"
)

func SetupDatabase(cfg config.Config) (*sql.DB, error) {
	dbURL := cfg.DBConnectionString
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}
	// run migrations or initial setup if needed
	return db, nil
}

type Record struct {
	ID       int    `json:"id"`
	Column1  string `json:"column1"`
	Column2  string `json:"column2"`
	Column3  string `json:"column3"`
	Column4  string `json:"column4"`
	Column5  string `json:"column5"`
	Column6  string `json:"column6"`
	Column7  string `json:"column7"`
	Column8  string `json:"column8"`
	Column9  string `json:"column9"`
	Column10 string `json:"column10"`
}

type Database struct {
	Connection *sql.DB
}

func NewDatabase(connStr string) (*Database, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &Database{Connection: db}, nil
}

func (d *Database) StoreRecords() ([]Record, error) {
	rows, err := d.Connection.Query("SELECT * FROM records")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		if err := rows.Scan(&record.ID, &record.Column1, &record.Column2, &record.Column3, &record.Column4, &record.Column5, &record.Column6, &record.Column7, &record.Column8, &record.Column9, &record.Column10); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (d *Database) CreateRecord(record Record) error {
	d.Connection.Exec(`
	       INSERT INTO files (
		       checksum, status, timestamp, original_filename, user_id, size, content_type, 
		       storage_path, submitter_ip, submitter_hostname, source_ip, source_hostname,
		       original_storage_path, upload_id
	       ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	       ON CONFLICT (checksum) DO NOTHING
       `, checksum, "stored", timestamp, input.OrigFilename, input.UserID, input.FileSize, input.ContentType,
		storagePath, input.SubmitterIP, input.SubmitterHostname, input.SourceIP, input.SourceHostname,
		input.OriginalStoragePath, input.UploadID)
	return err
}

func (d *Database) Close() error {
	return d.Connection.Close()
}
