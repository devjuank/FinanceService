# Finance AI Backend

This is the backend service for Finance AI, built with Go.

## Technologies
- **Go 1.20+**
- **JWT (json-web-token)** for secure authentication.
- **Bcrypt** for password hashing.
- **Native Go Normalization**: High-performance processing of PDF, CSV, and XLSX files.

## Project Structure
- `cmd/server/`: Main application entry point.
- `cmd/processor/`: Standalone CLI for manual data normalization.
- `internal/api/`: API handlers and middleware.
- `internal/auth/`: Authentication logic and JWT helpers.
- `internal/db/`: Data access layer (PostgreSQL) with Batch & Transaction support.
- `internal/models/`: Shared entities: **User**, **Transaction**, and **Upload** (Batches).
- `internal/processor/`: Core normalization engine and native parsers.
  - `parsers/`: Logic for Brubank, MercadoPago, Deel, and Santander.
  - `common/`: Shared helpers and ID generation logic.
  - `neutralizer.go`: Internal transfer matching logic.

## Running the Backend
```bash
go run cmd/server/main.go
```
The server will start on `http://localhost:8080`.
