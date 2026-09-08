// Command stuk-authkeys prints the SSH public keys stuk currently has an active
// grant for. It is meant for sshd's AuthorizedKeysCommand, so a key is accepted
// only while a grant is live (in addition to the firewall gate):
//
//	# /etc/ssh/sshd_config
//	AuthorizedKeysCommand      /usr/local/bin/stuk-authkeys -dir /run/stuk/keys
//	AuthorizedKeysCommandUser  root
//
// sshd appends its own arguments (user, key type, fingerprint, ...); they are
// ignored — grants are already IP-scoped and time-bounded by stukd.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/trustsentinel/stuk/internal/grant"
)

func main() {
	dir := flag.String("dir", "/run/stuk/keys", "directory of currently-granted keys")
	flag.Parse()

	keys, err := grant.AuthorizedKeys(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "stuk-authkeys:", err)
		os.Exit(1)
	}
	for _, k := range keys {
		fmt.Println(k)
	}
}
