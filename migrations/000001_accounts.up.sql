-- Identity core.
--
-- `subject` is what leaves this service inside a token. It is opaque and stable
-- and is deliberately NOT the primary key: a subject that is a row count leaks
-- how many users exist and invites callers to guess neighboring accounts.
CREATE TABLE users (
    id           BIGSERIAL PRIMARY KEY,
    subject      TEXT UNIQUE NOT NULL,
    display_name TEXT,
    active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_login_at TIMESTAMPTZ
);

-- Identifiers are a table, not a column: an account supports several verified
-- emails and phone numbers, and the primary one can be swapped.
CREATE TABLE identifiers (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind        TEXT NOT NULL CHECK (kind IN ('email', 'phone')),
    value       TEXT NOT NULL,          -- already normalized (lowercased, trimmed)
    verified_at TIMESTAMPTZ,
    is_primary  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (kind, value)
);

CREATE INDEX idx_identifiers_user ON identifiers (user_id);

-- One row per enrolled passkey.
CREATE TABLE passkey_credentials (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id BYTEA UNIQUE NOT NULL,  -- raw bytes; the assertion lookup key
    public_key    BYTEA NOT NULL,         -- COSE-encoded
    user_handle   BYTEA NOT NULL,         -- random 16 bytes, never derived from an identifier
    aaguid        UUID,
    sign_count    BIGINT NOT NULL DEFAULT 0,
    device_name   TEXT,

    -- Backup Eligible is IMMUTABLE after registration: go-webauthn compares the
    -- stored value against every assertion. Writing it as FALSE for a synced
    -- passkey lets the credential enroll and then fail every login, which is the
    -- single most expensive bug in this problem space.
    backup_eligible BOOLEAN NOT NULL,
    -- Backup State is mutable and refreshed on every successful login.
    backup_state    BOOLEAN NOT NULL,

    -- The RP ID this credential was created under. A passkey cannot move
    -- domains, so a future rename needs to find exactly the stale credentials
    -- and prompt those users to re-enroll.
    rp_id         TEXT NOT NULL,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at  TIMESTAMPTZ
);

CREATE INDEX idx_credentials_user ON passkey_credentials (user_id);

-- Short-lived rows for an email mid-registration or mid-recovery.
CREATE TABLE pending_registrations (
    id         BIGSERIAL PRIMARY KEY,
    email      TEXT NOT NULL,
    token      TEXT UNIQUE NOT NULL,
    code       TEXT,                     -- the 6-digit OTP, when that path is used
    attempts   INT NOT NULL DEFAULT 0,   -- a bounded guess count; an unbounded OTP is guessable
    purpose    TEXT NOT NULL CHECK (purpose IN ('registration', 'recovery')),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_pending_registrations_expires ON pending_registrations (expires_at);

-- WebAuthn challenges live in Postgres, not in process memory: an in-memory map
-- pins the service to a single replica, and this is the one component every
-- app depends on.
CREATE TABLE webauthn_challenges (
    id                         BIGSERIAL PRIMARY KEY,
    challenge                  BYTEA UNIQUE NOT NULL,
    purpose                    TEXT NOT NULL
                               CHECK (purpose IN ('registration', 'authentication', 'recovery')),
    user_id                    BIGINT REFERENCES users(id) ON DELETE CASCADE,
    pending_registration_token TEXT REFERENCES pending_registrations(token) ON DELETE CASCADE,
    session_data               BYTEA NOT NULL,   -- go-webauthn SessionData, JSON
    expires_at                 TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_challenges_expires ON webauthn_challenges (expires_at);

-- The single sign-on session: "this browser has a signed-in user". Opaque and
-- database-backed so revocation is immediate.
CREATE TABLE sso_sessions (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token        TEXT UNIQUE NOT NULL,
    user_agent   TEXT,
    ip           INET,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_sso_sessions_user ON sso_sessions (user_id);
