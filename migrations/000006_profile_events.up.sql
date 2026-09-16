-- A version for the profile, exposed as the standard OIDC updated_at claim
-- (Unix seconds), so a site can tell a newer profile from an older one.
ALTER TABLE users ADD COLUMN profile_updated_at BIGINT NOT NULL DEFAULT extract(epoch FROM now())::bigint;

-- Where a client receives pushed security events (RFC 8935), such as a profile
-- change. NULL means the client is not told.
ALTER TABLE oauth_clients ADD COLUMN events_uri TEXT;
