CREATE TABLE IF NOT EXISTS users (
			username TEXT NOT NULL PRIMARY KEY,
			displayname TEXT NOT NULL,
			email TEXT NOT NULL,
			password TEXT NOT NULL
		) WITHOUT ROWID;

CREATE UNIQUE INDEX IF NOT EXISTS users_idx ON users (username);
