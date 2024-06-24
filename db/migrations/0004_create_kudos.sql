CREATE TABLE IF NOT EXISTS kudos (
			id TEXT NOT NULL PRIMARY KEY,
			item_id TEXT NOT NULL,
			creator_username TEXT NOT NULL,
			emoji INTEGER NOT NULL,
			body TEXT NOT NULL,
			FOREIGN KEY (creator_username) REFERENCES users (username)
		) WITHOUT ROWID;

CREATE UNIQUE INDEX IF NOT EXISTS kudos_idx ON kudos (id);
CREATE INDEX IF NOT EXISTS kudos_item_idx ON kudos (item_id);
CREATE INDEX IF NOT EXISTS kudos_creator_usernamex ON kudos (creator_username);
