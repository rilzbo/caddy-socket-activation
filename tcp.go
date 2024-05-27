package caddy_socket_activation

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterNetwork("socket-activation", listenTCP)
}

func listenTCP(ctx context.Context, network, addr string, cfg net.ListenConfig) (any, error) {
	caddy.Log().Debug("Listen", zap.String("addr", addr))

	files := Files(true)
	if len(files) == 0 {
		return nil, errors.New("no file descriptors passed")
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	for _, file := range files {
		if ln, err := net.FileListener(file); err == nil {
			if file.Name() == host {
				return ln, nil
			} else {
				_ = ln.Close()
			}
		}
	}
	return nil, fmt.Errorf("no matching file descriptor found for %q", host)
}
