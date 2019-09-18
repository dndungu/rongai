package main

import (
	"log"
	"os"

	"github.com/dndungu/rongai/pkg/client"
)

func run() error {
	var c = &config{}

	if err := c.parse(os.Args); err != nil {
		return err
	}
	return client.New(
		client.WithHosts(c.hosts.Slice()...),
		client.WithKey(c.key),
		client.WithScript(c.script),
		client.WithPassphrase(c.passphrase),
		client.WithUser(c.user.String()),
		client.WithTimeout(c.timeout),
	).Run()
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}
