package caddy_socket_activation

import (
	"context"
	"errors"
	"net"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddyhttp.RegisterNetworkHTTP3("socket-activation", "socket-activation-udp")
	caddy.RegisterNetwork("socket-activation-udp", listenUDP)
}

func listenUDP(ctx context.Context, network, addr string, cfg net.ListenConfig) (any, error) {
	files := Files(true)
	if len(files) == 0 {
		return nil, errors.New("no file descriptors passed")
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	for _, file := range files {
		if pc, err := net.FilePacketConn(file); err == nil {
			if file.Name() == host {
				return pc, nil
			} else {
				_ = pc.Close()
			}
		}
	}
	return nil, errors.New("no matching file descriptor found")
}
