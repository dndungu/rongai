package client

import (
	"io"
	"time"
)

type config struct {
	hosts                   []string
	key, passphrase, script []byte
	stdin                   io.Reader
	stderr, stdout          io.Writer
	timeout                 time.Duration
	user                    string
}

type Option func(*client)

func WithHosts(hosts ...string) Option {
	return func(c *client) {
		c.config.hosts = hosts
	}
}

func WithKey(k []byte) Option {
	return func(c *client) {
		c.config.key = k
	}
}

func WithScript(s []byte) Option {
	return func(c *client) {
		c.config.script = s
	}
}

func WithPassphrase(p []byte) Option {
	return func(c *client) {
		c.config.passphrase = p
	}
}

func WithUser(u string) Option {
	return func(c *client) {
		c.config.user = u
	}
}

func WithTimeout(d time.Duration) Option {
	return func(c *client) {
		c.config.timeout = d
	}
}
