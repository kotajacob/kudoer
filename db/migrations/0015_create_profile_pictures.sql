CREATE TABLE IF NOT EXISTS profile_pictures (
			filename TEXT NOT NULL,
			username TEXT NOT NULL,
			kind INTEGER NOT NULL,
			CONSTRAINT username_key PRIMARY KEY (username, kind),
			FOREIGN KEY (username) REFERENCES users (username)
		) WITHOUT ROWID;

CREATE UNIQUE INDEX IF NOT EXISTS profile_pictures_filenamex ON profile_pictures (filename);
CREATE INDEX IF NOT EXISTS profile_pictures_usernamex ON profile_pictures (username);
CREATE INDEX IF NOT EXISTS profile_pictures_kindx ON profile_pictures (kind);
