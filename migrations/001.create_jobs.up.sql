CREATE TABLE jobs (
    job_id TEXT PRIMARY KEY,
    status TEXT NOT NULL,
    retries INTEGER NOT NULL,
    lifetime INTEGER NOT NULL,
    attempts INTEGER NOT NULL DEFAULT 0,
    duration_sec INTEGER NOT NULL DEFAULT 0
);