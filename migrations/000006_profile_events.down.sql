ALTER TABLE oauth_clients DROP COLUMN IF EXISTS events_uri;
ALTER TABLE users DROP COLUMN IF EXISTS profile_updated_at;
