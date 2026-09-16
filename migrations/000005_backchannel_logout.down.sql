ALTER TABLE oauth_clients DROP COLUMN IF EXISTS backchannel_logout_uri;
DROP INDEX IF EXISTS idx_refresh_tokens_user_live;
ALTER TABLE refresh_tokens      DROP COLUMN IF EXISTS sid;
ALTER TABLE authorization_codes DROP COLUMN IF EXISTS sid;
DROP INDEX IF EXISTS idx_sso_sessions_sid;
ALTER TABLE sso_sessions DROP COLUMN IF EXISTS sid;
