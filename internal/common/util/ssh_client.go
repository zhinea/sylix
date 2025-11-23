package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	client *ssh.Client
}

func NewSSHClient(host string, port int, user string, password *string, key *string) (*SSHClient, error) {
	var authMethods []ssh.AuthMethod

	if password != nil && *password != "" {
		authMethods = append(authMethods, ssh.Password(*password))
	}

	if key != nil && *key != "" {
		signer, err := ssh.ParsePrivateKey([]byte(*key))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method provided")
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: In production, use known_hosts
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return &SSHClient{client: client}, nil
}

func (s *SSHClient) Close() error {
	return s.client.Close()
}

func (s *SSHClient) RunCommand(cmd string) (string, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("failed to run command: %s, stderr: %s, err: %w", cmd, stderr.String(), err)
	}

	return stdout.String(), nil
}

func (s *SSHClient) CopyFile(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	w, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	var stderr bytes.Buffer
	session.Stderr = &stderr

	go func() {
		defer w.Close()
		io.Copy(w, srcFile)
	}()

	if err := session.Run("cat > " + dstPath); err != nil {
		return fmt.Errorf("failed to copy file: %w, stderr: %s", err, stderr.String())
	}

	return nil
}
