package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

type mandatory struct {
	name, value string
}

func (m *mandatory) Slice() []string {
	s := strings.Split(m.value, ",")
	for i := range s {
		s[i] = strings.TrimSpace(s[i])
	}
	return s
}

func (m *mandatory) String() string {
	return m.value
}

func (m *mandatory) Set(s string) error {
	m.value = s
	return nil
}

type config struct {
	fs                        *flag.FlagSet
	hosts                     *mandatory
	scriptfile, keyfile, user *mandatory
	key, passphrase, script   []byte
	dryrun, prompt            bool
	stdin                     io.Reader
	stderr, stdout            io.Writer
	timeout                   time.Duration
}

const defaultTimeout = 5 * time.Second

func (c *config) parse(args []string) error {
	c.hosts = &mandatory{"hosts", ""}
	c.scriptfile = &mandatory{"script", ""}
	c.keyfile = &mandatory{"key", ""}
	c.user = &mandatory{"user", ""}

	c.fs = flag.NewFlagSet(args[0], flag.ExitOnError)

	c.fs.BoolVar(
		&c.dryrun,
		"dryrun",
		c.dryrun,
		"Dry run script by prepending it with `set -nv`",
	)

	c.fs.Var(
		c.scriptfile,
		c.scriptfile.name,
		"Path to file with bash script to run remotely.",
	)

	c.fs.Var(
		c.hosts,
		c.hosts.name,
		"Comma separated list of remotes host:port or IP:PORT.",
	)

	c.fs.Var(
		c.keyfile,
		c.keyfile.name,
		"Path to file SSH key to authenticate with.",
	)

	c.fs.BoolVar(
		&c.prompt,
		"prompt",
		c.prompt,
		"Prompt for passphrase to parse SSH key.",
	)

	c.fs.DurationVar(
		&c.timeout,
		"timeout",
		defaultTimeout,
		"Timeout to connect.",
	)

	c.fs.Var(
		c.user,
		c.user.name,
		"User to authenticate as.",
	)

	var err error
	if err = c.fs.Parse(args[1:]); err != nil {
		return err
	}

	if c.script, err = ioutil.ReadFile(c.scriptfile.String()); err != nil {
		return err
	}

	if c.dryrun {
		c.script = append([]byte("set -nv\n"), c.script...)
	}

	if c.key, err = ioutil.ReadFile(c.keyfile.String()); err != nil {
		return err
	}

	if c.prompt {
		fmt.Fprint(os.Stdout, "SSH passphrase: ")
		if c.passphrase, err = terminal.ReadPassword(syscall.Stdin); err != nil {
			return err
		}
	}

	return nil
}
