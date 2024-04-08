DROP TABLE IF EXISTS su.public.accounts;

CREATE TABLE IF NOT EXISTS su.public.accounts
(
    id      BIGSERIAL PRIMARY KEY,
    balance INT NOT NULL DEFAULT 0
);
