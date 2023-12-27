BEGIN TRANSACTION;

ALTER TABLE urls
    ADD COLUMN IF NOT EXISTS userid VARCHAR(30) NOT NULL DEFAULT '_',
    ADD CONSTRAINT userid_not_empty CHECK (userid != '');

CREATE INDEX urls_userid ON urls (userid);

COMMIT;
