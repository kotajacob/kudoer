CREATE TABLE IF NOT EXISTS items (
			id TEXT NOT NULL PRIMARY KEY,
			creator_username TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			image TEXT NOT NULL,
			FOREIGN KEY (creator_username) REFERENCES users (username)
		) WITHOUT ROWID;

CREATE UNIQUE INDEX IF NOT EXISTS items_idx ON items (id);
