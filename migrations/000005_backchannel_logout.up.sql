-- A public identifier for each SSO session, carried as the "sid" claim, so a
-- site can be told which of its sessions a sign-out ends. The cookie token
-- stays secret; sid is safe to hand to every site.
ALTER TABLE sso_sessions ADD COLUMN sid TEXT;
UPDATE sso_sessions SET sid = replace(gen_random_uuid()::text, '-', '') WHERE sid IS NULL;
ALTER TABLE sso_sessions ALTER COLUMN sid SET NOT NULL;
CREATE UNIQUE INDEX idx_sso_sessions_sid ON sso_sessions (sid);

-- Which SSO session a code or refresh token was issued under, so signing that
-- browser out revokes exactly its tokens and notifies exactly its sites. NULL
-- on rows issued before this migration.
ALTER TABLE authorization_codes ADD COLUMN sid TEXT;
ALTER TABLE refresh_tokens      ADD COLUMN sid TEXT;
CREATE INDEX idx_refresh_tokens_user_live ON refresh_tokens (user_id) WHERE revoked_at IS NULL;

-- Where a client receives OpenID Connect back-channel logout tokens. NULL means
-- the client is not told when a user signs out.
ALTER TABLE oauth_clients ADD COLUMN backchannel_logout_uri TEXT;
