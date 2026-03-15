package ssh

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gossh "golang.org/x/crypto/ssh"

	"gluon-chaosmonkey/internal/config"
)

type Client struct {
	conn *gossh.Client
}

func NewClient(node config.NodeConfig, user, keyPath string) (*Client, error) {
	keyPath = expandHome(keyPath)

	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("read key %s: %w", keyPath, err)
	}

	signer, err := gossh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse key %s: %w", keyPath, err)
	}

	cfg := &gossh.ClientConfig{
		User: user,
		Auth: []gossh.AuthMethod{
			gossh.PublicKeys(signer),
		},
		// Accept any host key — for a chaos-testing tool in a lab environment
		// this is an acceptable trade-off; swap for a known_hosts verifier in prod.
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), //nolint:gosec
	}

	addr := node.Host + ":22"
	conn, err := gossh.Dial("tcp", addr, cfg)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "unable to authenticate") ||
			strings.Contains(msg, "no supported methods remain") ||
			strings.Contains(msg, "Permission denied") {
			return nil, fmt.Errorf("auth failed connecting to %s: %w", addr, err)
		}
		return nil, fmt.Errorf("unreachable %s: %w", addr, err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Run(cmd string) (stdout, stderr string, err error) {
	sess, err := c.conn.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("new session: %w", err)
	}
	defer sess.Close()

	var outBuf, errBuf bytes.Buffer
	sess.Stdout = &outBuf
	sess.Stderr = &errBuf

	runErr := sess.Run(cmd)
	return outBuf.String(), errBuf.String(), runErr
}

func (c *Client) RunBackground(cmd string) error {
	sess, err := c.conn.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}
	defer sess.Close()

	wrapped := fmt.Sprintf("nohup sh -c %q </dev/null >/dev/null 2>&1 &", cmd)
	if err := sess.Start(wrapped); err != nil {
		return fmt.Errorf("start background command: %w", err)
	}
	// Do NOT call sess.Wait() — we want fire-and-forget behaviour.
	return nil
}

func expandHome(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}
