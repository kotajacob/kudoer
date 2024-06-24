CREATE VIRTUAL TABLE IF NOT EXISTS items_search USING fts5(
			id,
			name,
			description
		);

CREATE TRIGGER IF NOT EXISTS after_items_insert AFTER INSERT ON items
			BEGIN INSERT INTO items_search(
				id,
				name,
				description
			)
			VALUES (
				new.id,
				new.name,
				new.description
			);
		END;
