SELECT :name, 'café', '$@#{x}<>&', '\\'
FROM records
WHERE id = :id;
