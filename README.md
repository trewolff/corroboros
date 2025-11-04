# corroboros
This Go service provides a simple REST API to create and query forensic file-presence records. It accepts file uploads, calculates strong cryptographic hashes, generates a cryptographic timestamp, and stores an encrypted, tamper-evident record in a database. Records are suitable for verifying that a specific file existed at a recorded time without exposing file contents.
Key features:

- File ingestion via multipart/form-data or binary upload
- Hashing with configurable algorithms (SHA-256 by default; optional SHA-512, BLAKE2)
- Canonicalization step (optional) to ensure consistent hashing of text files
- Cryptographic timestamping (RFC 3161-compatible or internally signed timestamp)
- Encrypted storage of records (AES-GCM or equivalent) with optional HSM/KMS integration
- Immutable, audit-friendly record format including hash, timestamp, metadata, and signature
- Verification endpoint to validate a file against a stored record (re-hash + timestamp/signature check)
- Access controls and logging for forensic chain-of-custody
- Optional export of verification artifacts (signed statements, timestamp receipts)
