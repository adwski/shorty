BEGIN TRANSACTION;

ALTER TABLE urls
    ADD COLUMN IF NOT EXISTS ts timestamp default current_timestamp;

CREATE INDEX urls_ts ON urls (ts);

COMMIT;
