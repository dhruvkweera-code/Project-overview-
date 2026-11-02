ALTER TABLE messages ADD COLUMN updated_at timestamptz NOT NULL DEFAULT now();

UPDATE messages SET updated_at = created_at;

-- Leave 10% free space in each page so updates can be done in place (HOT updates).
ALTER TABLE messages SET (fillfactor = 90);
