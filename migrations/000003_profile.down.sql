-- Display names cleared by the up migration are not restored: they were email
-- addresses, and putting them back would undo the point of it.
ALTER TABLE users DROP COLUMN IF EXISTS avatar;
