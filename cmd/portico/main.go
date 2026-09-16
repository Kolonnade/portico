// Command portico runs a Portico identity provider with the built-in extension
// points: a display name and emoji avatar for the profile, and email by SMTP.
package main

import (
	"os"

	"github.com/Kolonnade/portico/provider"
	"github.com/Kolonnade/portico/server"
)

func main() {
	os.Exit(server.Main("portico", os.Args[1:], provider.Options{}))
}
