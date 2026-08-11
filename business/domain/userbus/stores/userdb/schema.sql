-- Schema for the user domain.

CREATE TABLE IF NOT EXISTS users (
	user_id      BIGSERIAL   NOT NULL,
	name         TEXT        NOT NULL,
	email        TEXT        NOT NULL UNIQUE,
	enabled      BOOLEAN     NOT NULL DEFAULT TRUE,
	date_created TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	date_updated TIMESTAMPTZ NOT NULL DEFAULT NOW(),

	PRIMARY KEY (user_id)
);
