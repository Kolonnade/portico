// Package profile holds what an account holder chooses to be shown as: a
// display name and one of a fixed set of emoji avatars.
//
// Neither is an identifier. Nothing authenticates or authorizes on them; they
// exist so an account can be presented — on the home page and in the passkey's
// user name — without presenting its email address.
package profile

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Avatar is one selectable avatar. Key is what is stored.
type Avatar struct {
	Key, Emoji, Label string
}

// Avatars is the complete choice, in display order.
//
// Storing the key rather than the emoji keeps the column meaningful if the
// rendering ever changes to images.
var Avatars = []Avatar{
	{"fox", "🦊", "Fox"},
	{"panda", "🐼", "Panda"},
	{"octopus", "🐙", "Octopus"},
	{"owl", "🦉", "Owl"},
	{"turtle", "🐢", "Turtle"},
	{"lion", "🦁", "Lion"},
	{"frog", "🐸", "Frog"},
	{"bee", "🐝", "Bee"},
	{"cactus", "🌵", "Cactus"},
	{"rocket", "🚀", "Rocket"},
}

// Placeholder is shown until an avatar is chosen.
const Placeholder = "👤"

// MaxNameLen is the longest display name, in characters.
const MaxNameLen = 40

var (
	ErrNameEmpty     = errors.New("display name is required")
	ErrNameTooLong   = errors.New("display name is limited to 40 characters")
	ErrNameInvalid   = errors.New("display name contains characters that are not allowed")
	ErrAvatarUnknown = errors.New("choose one of the avatars")
)

// ValidAvatar reports whether key is one of Avatars.
func ValidAvatar(key string) bool {
	for _, a := range Avatars {
		if a.Key == key {
			return true
		}
	}
	return false
}

// EmojiFor renders a stored key, falling back to Placeholder.
func EmojiFor(key string) string {
	for _, a := range Avatars {
		if a.Key == key {
			return a.Emoji
		}
	}
	return Placeholder
}

// NormalizeName trims and collapses whitespace, then validates.
//
// Bidirectional overrides and isolates are refused because the name is shown
// to the holder in password managers and account pickers, where a reversed
// run of text is a way to make one name look like another.
func NormalizeName(s string) (string, error) {
	if !utf8.ValidString(s) {
		return "", ErrNameInvalid
	}
	name := strings.Join(strings.Fields(s), " ")
	if name == "" {
		return "", ErrNameEmpty
	}
	for _, r := range name {
		if unicode.IsControl(r) || isBidiControl(r) {
			return "", ErrNameInvalid
		}
	}
	if utf8.RuneCountInString(name) > MaxNameLen {
		return "", ErrNameTooLong
	}
	return name, nil
}

func isBidiControl(r rune) bool {
	return r == '‎' || r == '‏' || // LRM, RLM
		(r >= '‪' && r <= '‮') || // embeddings and overrides
		(r >= '⁦' && r <= '⁩') // isolates
}
