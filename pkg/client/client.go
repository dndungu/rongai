package client

import (
	"io"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

type client struct {
	config       config
	clientConfig *ssh.ClientConfig
	clients      []*ssh.Client
	signer       ssh.Signer
}

func (c *client) parseKey() error {
	var err error
	if len(c.config.passphrase) > 0 {
		c.signer, err = ssh.ParsePrivateKeyWithPassphrase(c.config.key, c.config.passphrase)
	} else {
		c.signer, err = ssh.ParsePrivateKey(c.config.key)
	}

	return err
}

const tcp = "tcp"

func (c *client) sshConfig() *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User:            c.config.user,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(c.signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
}
func (c *client) dial() error {
	var E = &Error{op: "client.dial"}
	var n = len(c.config.hosts)
	if n == 0 {
		return E.Msg("no hosts provided")
	}

	for i := 0; i < n; i++ {
		var err = func(i int) error {
			var (
				conn net.Conn
				err  error
			)

			if i == 0 {
				conn, err = net.DialTimeout(tcp, c.config.hosts[i], c.config.timeout)
			} else {
				conn, err = c.clients[i-1].Dial(tcp, c.config.hosts[i])
			}

			if err != nil {
				return E.Err(err)
			}

			cc, chans, reqs, err := ssh.NewClientConn(conn, c.config.hosts[i], c.sshConfig())
			if err != nil {
				return E.Err(err)
			}

			c.clients = append(c.clients, ssh.NewClient(cc, chans, reqs))

			return nil
		}(i)

		if err != nil {
			return E.Err(err)
		}
	}

	return nil
}

func (c *client) client() *ssh.Client {
	var n = len(c.clients)
	if n == 0 {
		return nil
	}
	return c.clients[n-1]
}

func (c *client) run() error {
	var err error
	var session *ssh.Session
	session, err = c.client().NewSession()
	if err != nil {
		return err
	}

	defer session.Close()

	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}

	go io.Copy(c.config.stderr, stderr)

	stdin, err := session.StdinPipe()
	if err != nil {
		return err
	}

	defer stdin.Close()

	go io.Copy(stdin, c.config.stdin)

	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	go io.Copy(c.config.stdout, stdout)

	var termModes = ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 40, 80, termModes); err != nil {
		return err
	}

	return session.Run(string(c.config.script))
}

func (c *client) Run() error {
	for _, fn := range []func() error{
		c.parseKey,
		c.dial,
		c.run,
	} {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

type Runner interface {
	Run() error
}

func New(options ...Option) Runner {
	var c client

	c.config.stdin, c.config.stderr, c.config.stdout = os.Stdin, os.Stderr, os.Stdout

	for _, o := range options {
		o(&c)
	}

	return &c
}
