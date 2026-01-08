ls# Finance AI

Finance AI is a premium financial intelligence platform that automates the normalization of bank statements and provides interactive visualizations for better financial management.

## Project Structure

- **`/backend`**: Go-based API server for user management and file processing orchestration.
- **`/frontend`**: Next.js application with a premium dark-mode dashboard and data visualizations.

---

## API Contract

The backend service runs at `http://localhost:8080` by default.

### Authentication

#### POST `/api/register`
Creates a new user account.
**Request Body:** `{"email": "...", "password": "..."}`

#### POST `/api/login`
Authenticates a user and returns a JWT.
**Request Body:** `{"email": "...", "password": "..."}`
**Response:** `{"token": "..."}`

### Financial Data

#### POST `/api/upload`
**Header:** `Authorization: Bearer <token>`
Uploads a bank statement (PDF, CSV, XLSX). The system creates an **Import Batch** (Upload) to track the origin of the data.
**Request Body:** `multipart/form-data` (field `file`)
**Response:** `{"upload_id": "...", "count": 12, "message": "..."}`

#### GET `/api/transactions`
**Header:** `Authorization: Bearer <token>`
Retrieves normalized transactions for the authenticated user. Includes `upload_id` for traceability.

#### GET `/api/uploads`
**Header:** `Authorization: Bearer <token>`
Lists all previous import batches.

---

## Getting Started

1. **Backend**: Follow instructions in [backend/README.md](./backend/README.md)
2. **Frontend**: Follow instructions in [frontend/README.md](./frontend/README.md)
