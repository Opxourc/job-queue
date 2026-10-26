# Job Queue

A small Go HTTP service for creating, tracking, and canceling background jobs. It is a learning project that demonstrates asynchronous work with goroutines and contexts; jobs simulate work rather than processing real tasks.

## Requirements

- Go 1.27 or later
- SQLite
- The `sqlite3` command-line tool to initialize the database (or another SQLite client)

## Getting started

1. Clone the repository and change into its directory.
2. Create the database directory and initialize the database using the included schema:

   ```sh
   mkdir -p data
   sqlite3 data/app.db ".read migrations/001.create_jobs.up.sql"
   ```

   On Windows PowerShell, create the directory with `New-Item -ItemType Directory -Force data`, then run the `sqlite3` command above.

3. Start the API:

   ```sh
   go run ./cmd/api
   ```

The server listens on port `8080`. It opens `data/app.db` relative to the project root. The application does not apply database migrations automatically; initialize the schema before starting it.

## API

All routes are under `/api/jobs`.

| Method | Path | Description | Success response |
| --- | --- | --- | --- |
| `POST` | `/api/jobs/` | Create and start a job | `202 Accepted` with the job ID |
| `GET` | `/api/jobs/:id` | Get one job by UUID | `200 OK` with a job object |
| `GET` | `/api/jobs/` | List all jobs | `200 OK` with an array of jobs |
| `POST` | `/api/jobs/:id/cancel` | Cancel a running job | `204 No Content` |

Invalid UUIDs and missing jobs return `400 Bad Request` and `404 Not Found`, respectively.

### Create a job

```sh
curl -X POST http://localhost:8080/api/jobs/ \
  -H "Content-Type: application/json" \
  -d '{"retries": 2, "lifetime": 30000000000}'
```

The request fields are:

- `retries`: maximum number of retries, clamped to `0`–`5`.
- `lifetime`: duration as an integer number of nanoseconds, clamped to `0`–`180000000000` (three minutes).

The response body is the new job's UUID. Use it to retrieve or cancel the job:

```sh
curl http://localhost:8080/api/jobs/JOB_UUID
curl -X POST http://localhost:8080/api/jobs/JOB_UUID/cancel
```

A job object contains `job_id`, `status`, `retries`, `lifetime`, `attempts`, and `duration_sec`. Its status progresses from `queued` to `running`, then to `succeeded`, `failed`, or `canceled`.

## Current behavior and limitations

Job execution is simulated: each attempt waits a random 1–10 seconds and succeeds or fails randomly. A job makes at most `retries + 1` attempts. The `lifetime` value is currently stored but is not used to limit execution time. Jobs are run in-process, so they do not resume after the server stops.

## Tests

Run the Go test suite with:

```sh
go test ./...
```
