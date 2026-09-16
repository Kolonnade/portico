package accounts

import (
	"encoding/json"

	"github.com/go-webauthn/webauthn/webauthn"
)

// encodeSession and decodeSession persist an in-flight ceremony.
//
// The library's own SessionData is marshalled whole, deliberately.
//
// An earlier version of this file kept a hand-written subset — challenge, user
// id, allowed credentials, user verification — on the theory that pinning the
// stored shape protected against library churn. It did the opposite: it dropped
// CredParams, which FinishRegistration uses to check the credential's algorithm
// against what was offered, so every enrollment failed with "Invalid attestation
// format". Storing the whole value means a field added upstream travels with it
// instead of being silently lost.
//
// The rows are short-lived (a five-minute ceremony), so there is no long-term
// compatibility burden in return.
func encodeSession(s *webauthn.SessionData) ([]byte, error) {
	return json.Marshal(s)
}

func decodeSession(blob []byte) (webauthn.SessionData, error) {
	var s webauthn.SessionData
	err := json.Unmarshal(blob, &s)
	return s, err
}
