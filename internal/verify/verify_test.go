package verify

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	jose "github.com/go-jose/go-jose/v4"
)

// F1: tokens carrying unknown key ids must not turn a relying party into a
// request amplifier aimed at the provider.
func TestUnknownKeyIDsDoNotAmplifyFetches(t *testing.T) {
	var fetches atomic.Int32
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fetches.Add(1)
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{
			{Key: &priv.PublicKey, KeyID: "real", Algorithm: "ES256", Use: "sig"},
		}})
	}))
	defer srv.Close()

	ks := NewKeySet(srv.URL, nil)
	ctx := context.Background()
	if _, err := ks.Key(ctx, "real"); err != nil {
		t.Fatalf("known key: %v", err)
	}
	for i := 0; i < 200; i++ {
		if _, err := ks.Key(ctx, fmt.Sprintf("forged-%d", i)); err == nil {
			t.Fatal("an unknown kid resolved to a key")
		}
	}
	if n := fetches.Load(); n != 1 {
		t.Fatalf("%d fetches for 200 unknown kids, want 1", n)
	}

	// A real rotation is still picked up once the interval has passed.
	ks.SetMinInterval(10 * time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	_, _ = ks.Key(ctx, "rotated")
	if n := fetches.Load(); n != 2 {
		t.Fatalf("after the interval: %d fetches, want 2", n)
	}
}

// A token must carry exactly one audience and the expected type.
func TestTokenChecksAudienceAndType(t *testing.T) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{
			{Key: &priv.PublicKey, KeyID: "k", Algorithm: "ES256", Use: "sig"},
		}})
	}))
	defer srv.Close()

	sign := func(typ string, aud any) string {
		signer, _ := jose.NewSigner(jose.SigningKey{Algorithm: jose.ES256, Key: &jose.JSONWebKey{Key: priv, KeyID: "k"}},
			(&jose.SignerOptions{}).WithType(jose.ContentType(typ)))
		payload, _ := json.Marshal(map[string]any{
			"iss": "https://issuer", "sub": "usr_1", "aud": aud, "exp": time.Now().Add(time.Minute).Unix(),
		})
		obj, _ := signer.Sign(payload)
		s, _ := obj.CompactSerialize()
		return s
	}
	opts := Options{Issuer: "https://issuer", Audience: "groups", Keys: NewKeySet(srv.URL, nil), Type: "logout+jwt"}
	ctx := context.Background()

	if _, err := Token(ctx, sign("logout+jwt", "groups"), opts); err != nil {
		t.Fatalf("valid token refused: %v", err)
	}
	if _, err := Token(ctx, sign("JWT", "groups"), opts); err == nil {
		t.Error("a token of the wrong type was accepted")
	}
	if _, err := Token(ctx, sign("logout+jwt", []string{"groups", "photos"}), opts); err == nil {
		t.Error("a token with two audiences was accepted")
	}
	if _, err := Token(ctx, sign("logout+jwt", "photos"), opts); err == nil {
		t.Error("a token for another audience was accepted")
	}
}
