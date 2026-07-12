-- Idempotency keys structure per Part 9.4

CREATE TABLE idempotency_records (
    idempotency_key UUID PRIMARY KEY,
    request_hash TEXT NOT NULL,      -- hash of the full request payload
    response_snapshot JSONB,         -- cached response, replayed on duplicate
    status TEXT NOT NULL CHECK (status IN ('IN_PROGRESS','COMPLETED','FAILED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT now() + interval '24 hours'
);

-- Index for quick purging of expired records
CREATE INDEX idx_idempotency_records_expires_at ON idempotency_records (expires_at);
