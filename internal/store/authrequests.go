package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AuthRequest is an authorization request held between /oauth/authorize and
// the moment a signed-in account completes it.
type AuthRequest struct {
	ID                  string
	ClientID            string
	RedirectURI         string
	ResponseType        string
	ResponseMode        string
	Scopes              []string
	State               string
	Nonce               string
	Prompt              []string
	MaxAge              *int
	LoginHint           string
	HintSubject         string
	CodeChallenge       string
	CodeChallengeMethod string

	// Set when the request is completed.
	UserID   *int64
	Subject  string
	SID      string
	AuthTime time.Time
	AMR      []string
	Done     bool

	CreatedAt time.Time
	ExpiresAt time.Time
}

// AuthRequestTTL bounds how long a person has to sign in.
const AuthRequestTTL = 10 * time.Minute

// AuthCodeTTL is deliberately short: the code is handed over in a URL, and its
// only job is to survive one redirect.
const AuthCodeTTL = 60 * time.Second

const authRequestColumns = `a.id, a.client_id, a.redirect_uri, a.response_type, a.response_mode, a.scopes,
	a.state, a.nonce, a.prompt, a.max_age, a.login_hint, a.hint_subject, a.code_challenge,
	a.code_challenge_method, a.user_id, COALESCE(u.subject, ''), COALESCE(a.sid, ''),
	COALESCE(a.auth_time, 'epoch'::timestamptz), a.amr, a.done, a.created_at, a.expires_at`

func scanAuthRequest(row interface{ Scan(...any) error }) (*AuthRequest, error) {
	var a AuthRequest
	if err := row.Scan(&a.ID, &a.ClientID, &a.RedirectURI, &a.ResponseType, &a.ResponseMode, &a.Scopes,
		&a.State, &a.Nonce, &a.Prompt, &a.MaxAge, &a.LoginHint, &a.HintSubject, &a.CodeChallenge,
		&a.CodeChallengeMethod, &a.UserID, &a.Subject, &a.SID, &a.AuthTime, &a.AMR, &a.Done,
		&a.CreatedAt, &a.ExpiresAt); err != nil {
		return nil, norm(err)
	}
	return &a, nil
}

// CreateAuthRequest stores a new request and returns it with its id set.
func (d *DB) CreateAuthRequest(ctx context.Context, a AuthRequest) (*AuthRequest, error) {
	a.ID = uuid.NewString()
	a.ExpiresAt = time.Now().Add(AuthRequestTTL)
	if a.Prompt == nil {
		a.Prompt = []string{}
	}
	_, err := d.pool.Exec(ctx,
		`INSERT INTO auth_requests
		   (id, client_id, redirect_uri, response_type, response_mode, scopes, state, nonce, prompt,
		    max_age, login_hint, hint_subject, code_challenge, code_challenge_method, expires_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		a.ID, a.ClientID, a.RedirectURI, a.ResponseType, a.ResponseMode, a.Scopes, a.State, a.Nonce,
		a.Prompt, a.MaxAge, a.LoginHint, a.HintSubject, a.CodeChallenge, a.CodeChallengeMethod, a.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// AuthRequestByID loads a request that has not expired.
func (d *DB) AuthRequestByID(ctx context.Context, id string) (*AuthRequest, error) {
	return scanAuthRequest(d.pool.QueryRow(ctx,
		`SELECT `+authRequestColumns+`
		   FROM auth_requests a LEFT JOIN users u ON u.id = a.user_id
		  WHERE a.id = $1 AND a.expires_at > now()`, id))
}

// SaveAuthCode records the code issued for a completed request, and shortens
// the request's life to the code's.
func (d *DB) SaveAuthCode(ctx context.Context, id, code string) error {
	tag, err := d.pool.Exec(ctx,
		`UPDATE auth_requests SET code_hash = $2, expires_at = now() + $3::interval
		  WHERE id = $1 AND done`, id, HashToken(code), AuthCodeTTL.String())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ConsumeAuthCode resolves a code to its request exactly once.
//
// The code is cleared in the same statement that reads it, so two concurrent
// redemptions cannot both succeed; the second finds nothing.
func (d *DB) ConsumeAuthCode(ctx context.Context, code string) (*AuthRequest, error) {
	return scanAuthRequest(d.pool.QueryRow(ctx,
		`WITH hit AS (
		   UPDATE auth_requests SET code_hash = NULL
		    WHERE code_hash = $1 AND expires_at > now() AND done
		   RETURNING *)
		 SELECT `+authRequestColumns+` FROM hit a LEFT JOIN users u ON u.id = a.user_id`,
		HashToken(code)))
}

// CompleteAuthRequest records which signed-in account completes a request.
func (d *DB) CompleteAuthRequest(ctx context.Context, id string, userID int64, sid string, authTime time.Time, amr []string) error {
	if amr == nil {
		amr = []string{}
	}
	tag, err := d.pool.Exec(ctx,
		`UPDATE auth_requests SET user_id = $2, sid = $3, auth_time = $4, amr = $5, done = TRUE
		  WHERE id = $1 AND expires_at > now() AND NOT done`,
		id, userID, sid, authTime, amr)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteAuthRequest removes a request once its tokens are issued.
func (d *DB) DeleteAuthRequest(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx, `DELETE FROM auth_requests WHERE id = $1`, id)
	return err
}
