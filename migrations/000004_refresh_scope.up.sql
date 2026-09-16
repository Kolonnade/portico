-- The scope a refresh token was granted, so a refresh issues tokens with the
-- same claims as the sign-in did. Refresh used to issue "openid" only, which
-- silently dropped the profile claims after one access-token lifetime.
ALTER TABLE refresh_tokens ADD COLUMN scope TEXT NOT NULL DEFAULT 'openid';

-- Every client registered before this migration requested "profile", so live
-- tokens are backfilled with it. "email" is deliberately not backfilled: the
-- sites no longer ask for it.
UPDATE refresh_tokens SET scope = 'openid profile' WHERE revoked_at IS NULL;
