CREATE TABLE IF NOT EXISTS users_following (
	username TEXT NOT NULL,
	following_username TEXT NOT NULL,
	CONSTRAINT follow_key PRIMARY KEY (username, following_username),
	FOREIGN KEY (username) REFERENCES users (username),
	FOREIGN KEY (following_username) REFERENCES users (username)
) WITHOUT ROWID;

CREATE INDEX IF NOT EXISTS users_following_username ON users_following (username);
CREATE INDEX IF NOT EXISTS users_following_following_username ON users_following (following_username);
