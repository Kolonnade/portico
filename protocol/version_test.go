package protocol

import "testing"

func TestParseVersion(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want Version
		ok   bool
	}{
		{"1.0", Version{1, 0}, true},
		{"2.13", Version{2, 13}, true},
		{" 1.0 ", Version{1, 0}, true},
		{"1", Version{}, false},
		{"", Version{}, false},
		{"x.y", Version{}, false},
		{"-1.0", Version{}, false},
	} {
		got, err := ParseVersion(tc.in)
		if tc.ok != (err == nil) {
			t.Fatalf("ParseVersion(%q): err=%v, want ok=%v", tc.in, err, tc.ok)
		}
		if tc.ok && got != tc.want {
			t.Fatalf("ParseVersion(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// A minor-version difference must never break a client: that is the whole
// promise of the versioning scheme, so it is asserted directly.
func TestNegotiateWithinMajor(t *testing.T) {
	server := []Version{{1, 0}, {1, 1}, {1, 2}}

	if got, err := Negotiate(Version{1, 5}, server); err != nil || got != (Version{1, 2}) {
		t.Fatalf("newer client: got %v, %v; want 1.2", got, err)
	}
	if got, err := Negotiate(Version{1, 0}, server); err != nil || got != (Version{1, 0}) {
		t.Fatalf("older client: got %v, %v; want 1.0", got, err)
	}
	if got, err := Negotiate(Version{1, 2}, server); err != nil || got != (Version{1, 2}) {
		t.Fatalf("exact match: got %v, %v; want 1.2", got, err)
	}
}

// A major mismatch must be a clean, informative error rather than a silent
// downgrade to something neither side agreed to.
func TestNegotiateAcrossMajorFails(t *testing.T) {
	server := []Version{{1, 0}, {1, 1}}
	_, err := Negotiate(Version{2, 0}, server)
	if err == nil {
		t.Fatal("expected an error negotiating 2.0 against a 1.x-only server")
	}
	var unsupported *ErrUnsupportedVersion
	if !asUnsupported(err, &unsupported) {
		t.Fatalf("expected ErrUnsupportedVersion, got %T", err)
	}
	if len(unsupported.Supported) != 2 {
		t.Fatalf("error should carry the supported list, got %v", unsupported.Supported)
	}
}

func asUnsupported(err error, target **ErrUnsupportedVersion) bool {
	e, ok := err.(*ErrUnsupportedVersion)
	if ok {
		*target = e
	}
	return ok
}

func TestPathPrefixUsesMajorOnly(t *testing.T) {
	if got := (Version{1, 7}).PathPrefix(); got != "/accounts/v1" {
		t.Fatalf("PathPrefix() = %q, want /accounts/v1", got)
	}
}
