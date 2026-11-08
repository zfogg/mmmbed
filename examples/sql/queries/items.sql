-- name: InsertItem
-- access: write
-- rows: one
INSERT INTO items (name)
VALUES (:name)
RETURNING id;

-- name: LookupItem
-- access: read
-- rows: zero-or-one
SELECT name
FROM items
WHERE id = :id;
