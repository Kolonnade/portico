package profile

import (
	"strings"
	"testing"
)

func TestNormalizeName(t *testing.T) {
	for _, tc := range []struct {
		in, want string
		err      error
	}{
		{"Brennan", "Brennan", nil},
		{"  Brennan \t  S\n", "Brennan S", nil},
		{"Zoë 🦉", "Zoë 🦉", nil},
		{strings.Repeat("é", MaxNameLen), strings.Repeat("é", MaxNameLen), nil},
		{"", "", ErrNameEmpty},
		{" \t\n ", "", ErrNameEmpty},
		{strings.Repeat("x", MaxNameLen+1), "", ErrNameTooLong},
		{"evil‮gnp.exe", "", ErrNameInvalid},
		{"nul\x00byte", "", ErrNameInvalid},
		{"bad\xffutf8", "", ErrNameInvalid},
	} {
		got, err := NormalizeName(tc.in)
		if err != tc.err || got != tc.want {
			t.Errorf("NormalizeName(%q) = %q, %v; want %q, %v", tc.in, got, err, tc.want, tc.err)
		}
	}
}

func TestAvatars(t *testing.T) {
	if len(Avatars) != 10 {
		t.Fatalf("want 10 avatars, have %d", len(Avatars))
	}
	seen := map[string]bool{}
	for _, a := range Avatars {
		if seen[a.Key] || seen[a.Emoji] {
			t.Errorf("duplicate avatar %q / %q", a.Key, a.Emoji)
		}
		seen[a.Key], seen[a.Emoji] = true, true
		if !ValidAvatar(a.Key) || EmojiFor(a.Key) != a.Emoji {
			t.Errorf("avatar %q does not round-trip", a.Key)
		}
	}
	if ValidAvatar("") || ValidAvatar("dragon") {
		t.Error("unknown keys must be invalid")
	}
	if EmojiFor("dragon") != Placeholder {
		t.Error("unknown key must render the placeholder")
	}
}
