-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "outbox" (
    id             UUID PRIMARY KEY,
    aggregate_type TEXT        NOT NULL,
    aggregate_id   UUID        NOT NULL,
    event_type     TEXT        NOT NULL,
    payload        JSONB       NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL,
    published_at   TIMESTAMPTZ
);
-- +goose StatementEnd

-- +goose StatementBegin
-- Partial index for the publisher worker: poll oldest-first among unpublished.
CREATE INDEX IF NOT EXISTS idx_outbox_unpublished
    ON "outbox" (created_at)
    WHERE published_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "outbox";
-- +goose StatementEnd
