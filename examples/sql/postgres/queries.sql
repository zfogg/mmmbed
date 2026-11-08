-- name: InsertItem
-- access: write
-- rows: one
INSERT INTO mmmbed_items (name) VALUES ($1)
RETURNING id;

-- name: LookupItem
-- access: read
-- rows: zero-or-one
SELECT name FROM mmmbed_items WHERE id = $1;

-- name: InspectLiterals
-- access: read
-- rows: one
SELECT ARRAY[']', 'a'] AS items, $тег$hello$тег$ AS label;

-- name: InspectEscapes
-- access: read
-- rows: one
SELECT E'first'
  'second\'value' AS result;
