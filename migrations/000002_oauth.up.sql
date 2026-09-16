-- Registered relying parties. Adding an app is a row, not a release.
CREATE TABLE oauth_clients (
    id                   BIGSERIAL PRIMARY KEY,
    client_id            TEXT UNIQUE NOT NULL,
    client_secret_hash   TEXT,             -- NULL for public clients using PKCE alone
    audience             TEXT NOT NULL,    -- what lands in `aud`; the app checks it matches itself
    display_name         TEXT NOT NULL,

    -- Exact-match only. No wildcards, no prefixes, no port flexibility: a loose
    -- redirect URI is how authorization codes get phished.
    redirect_uris        TEXT[] NOT NULL,

    -- first_party skips the consent screen. A server we do not operate must not
    -- be able to borrow our sign-in UI and look like us.
    trust_tier           TEXT NOT NULL DEFAULT 'third_party'
                         CHECK (trust_tier IN ('first_party', 'third_party')),

    -- Set once the operator proves control of the redirect URI's domain.
    domain_verified_at   TIMESTAMPTZ,
    domain_verify_token  TEXT,

    protocol_version_min TEXT NOT NULL DEFAULT '1.0',
    active               BOOLEAN NOT NULL DEFAULT TRUE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Authorization codes: single use, 60-second life, bound to a PKCE challenge.
CREATE TABLE authorization_codes (
    id             BIGSERIAL PRIMARY KEY,
    code_hash      TEXT UNIQUE NOT NULL,   -- hashed: a leaked table must not be redeemable
    client_id      TEXT NOT NULL REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
    user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    redirect_uri   TEXT NOT NULL,
    code_challenge TEXT NOT NULL,
    nonce          TEXT,
    scope          TEXT NOT NULL DEFAULT 'openid',
    expires_at     TIMESTAMPTZ NOT NULL,
    consumed_at    TIMESTAMPTZ             -- kept after use, so a replay is detectable
);

CREATE INDEX idx_auth_codes_expires ON authorization_codes (expires_at);

-- Rotating refresh tokens. `rotated_from` makes reuse of a superseded token
-- detectable, which is the signal that one has been stolen.
CREATE TABLE refresh_tokens (
    id           BIGSERIAL PRIMARY KEY,
    token_hash   TEXT UNIQUE NOT NULL,
    client_id    TEXT NOT NULL REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rotated_from TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ NOT NULL,
    revoked_at   TIMESTAMPTZ
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id);

-- Signing keys, with the state machine rotation needs.
--
-- pending  → published in the JWKS, not yet signing
-- active   → signing (exactly one at a time)
-- retiring → no longer signing, still published so outstanding tokens verify
-- retired  → removed from the JWKS
CREATE TABLE signing_keys (
    id             BIGSERIAL PRIMARY KEY,
    kid            TEXT UNIQUE NOT NULL,
    algorithm      TEXT NOT NULL DEFAULT 'ES256',
    private_key_pem BYTEA NOT NULL,
    public_key_jwk  JSONB NOT NULL,
    state          TEXT NOT NULL CHECK (state IN ('pending', 'active', 'retiring', 'retired')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    activated_at   TIMESTAMPTZ,
    retired_at     TIMESTAMPTZ
);

-- Only one key signs at a time.
CREATE UNIQUE INDEX idx_signing_keys_one_active ON signing_keys (state) WHERE state = 'active';

CREATE TABLE auth_audit (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT REFERENCES users(id) ON DELETE SET NULL,
    client_id  TEXT,
    action     TEXT NOT NULL,
    ip         INET,
    user_agent TEXT,
    detail     JSONB,
    at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_auth_audit_user_at ON auth_audit (user_id, at DESC);
