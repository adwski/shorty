BEGIN TRANSACTION;

CREATE TABLE urls (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    hash VARCHAR(20) NOT NULL,
    orig VARCHAR(500) NOT NULL UNIQUE,
    CONSTRAINT hash_not_empty CHECK (hash != ''),
    CONSTRAINT orig_not_empty CHECK (orig != '')
);

CREATE UNIQUE INDEX urls_hash On urls (hash);

COMMIT;
