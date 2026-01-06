# Finance AI Backend

This is the backend service for Finance AI, built with Go.

## Technologies
- **Go 1.20+**
- **JWT (json-web-token)** for secure authentication.
- **Bcrypt** for password hashing.
- **Python Integration** for file normalization logic.

## Project Structure
- `cmd/server/`: Main application entry point.
- `internal/api/`: API handlers and middleware.
- `internal/auth/`: Authentication logic and JWT helpers.
- `internal/db/`: Data access layer (currently in-memory, ready for PostgreSQL).
- `internal/models/`: Shared data structures (User, Transaction).
- `internal/processor/`: Orchestrator that triggers Python normalization scripts.

## Running the Backend
```bash
go run cmd/server/main.go
```
The server will start on `http://localhost:8080`.
