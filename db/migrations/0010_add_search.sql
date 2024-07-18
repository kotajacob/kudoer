CREATE VIRTUAL TABLE IF NOT EXISTS items_search USING fts5(
			id,
			name,
			tokenize = porter
		);

INSERT INTO items_search (id, name)
		SELECT
			id,
			name
		FROM
			items;

CREATE TRIGGER IF NOT EXISTS after_items_insert AFTER INSERT ON items
			BEGIN INSERT INTO items_search(
				id,
				name
			)
			VALUES (
				new.id,
				new.name
			);
		END;

CREATE VIRTUAL TABLE IF NOT EXISTS users_search USING fts5(
			username,
			displayname
		);

INSERT INTO users_search (username, displayname)
		SELECT
			username,
			displayname
		FROM
			users;

CREATE TRIGGER IF NOT EXISTS after_users_insert AFTER INSERT ON users
			BEGIN INSERT INTO users_search(
				username,
				displayname
			)
			VALUES (
				new.username,
				new.displayname
			);
		END;
