ALTER TABLE messages RESET (fillfactor);

ALTER TABLE messages DROP COLUMN updated_at;
