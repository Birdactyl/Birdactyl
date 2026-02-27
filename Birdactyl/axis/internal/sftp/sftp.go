package sftp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"cauthon-axis/internal/config"
	"cauthon-axis/internal/logger"
	"cauthon-axis/internal/panel"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var (
	server   net.Listener
	serverMu sync.Mutex
)

func Start(port int) error {
	serverMu.Lock()
	defer serverMu.Unlock()

	if server != nil {
		return nil
	}

	hostKey, err := loadOrGenerateHostKey()
	if err != nil {
		return fmt.Errorf("failed to load host key: %w", err)
	}

	sshConfig := &ssh.ServerConfig{
		PasswordCallback: authenticateUser,
	}
	sshConfig.AddHostKey(hostKey)

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	server = listener
	logger.Success("SFTP server listening on %s", addr)

	go acceptConnections(listener, sshConfig)

	return nil
}

func Stop() {
	serverMu.Lock()
	defer serverMu.Unlock()

	if server != nil {
		server.Close()
		server = nil
	}
}

func acceptConnections(listener net.Listener, sshConfig *ssh.ServerConfig) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			logger.Warn("SFTP accept error: %v", err)
			continue
		}

		go handleConnection(conn, sshConfig)
	}
}

func handleConnection(conn net.Conn, sshConfig *ssh.ServerConfig) {
	defer conn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(conn, sshConfig)
	if err != nil {
		return
	}
	defer sshConn.Close()

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}

		go handleRequests(requests)
		go handleSFTP(channel, sshConn.User())
	}
}

func handleRequests(requests <-chan *ssh.Request) {
	for req := range requests {
		ok := false
		switch req.Type {
		case "subsystem":
			if string(req.Payload[4:]) == "sftp" {
				ok = true
			}
		}
		req.Reply(ok, nil)
	}
}

func handleSFTP(channel ssh.Channel, username string) {
	defer channel.Close()

	parts := strings.Split(username, ".")
	if len(parts) < 2 {
		return
	}
	serverID := parts[len(parts)-1]

	cfg := config.Get()
	rootPath := filepath.Join(cfg.Node.DataDir, serverID)

	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return
	}

	sftpServer, err := sftp.NewServer(
		channel,
		sftp.WithServerWorkingDirectory(rootPath),
	)
	if err != nil {
		return
	}

	if err := sftpServer.Serve(); err != nil && err != io.EOF {
		logger.Warn("SFTP session error for %s: %v", serverID, err)
	}
}

func authenticateUser(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	username := conn.User()

	parts := strings.Split(username, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid username format")
	}
	serverID := parts[len(parts)-1]

	client := panel.NewClient()
	if err := client.ValidateSFTPCredentials(serverID, string(password)); err != nil {
		return nil, fmt.Errorf("authentication failed")
	}

	return &ssh.Permissions{
		Extensions: map[string]string{
			"server_id": serverID,
		},
	}, nil
}

func loadOrGenerateHostKey() (ssh.Signer, error) {
	keyPath := "sftp_host_key"

	if data, err := os.ReadFile(keyPath); err == nil {
		return ssh.ParsePrivateKey(data)
	}

	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		return nil, err
	}

	return ssh.ParsePrivateKey(keyPEM)
}
