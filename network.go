package caddy_network_fd

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

const (
	SD_LISTEN_FDS_START = 3
)

type SocketFilePredicate func(*SocketFile) bool

type SocketFile struct {
	sofamily int
	sotype   int
	file     *os.File
}

var fs []*SocketFile

func init() {
	caddy.RegisterNetwork("fd", listenStream)
	caddy.RegisterNetwork("fd-dgram", listenDGRAM)
	caddyhttp.RegisterNetworkHTTP3("fd", "fd-dgram")

	fs = files()
}

func isSocket(fd int) bool {
	var stat syscall.Stat_t
	err := syscall.Fstat(fd, &stat)
	if err == nil {
		return syscall.S_IFSOCK == (stat.Mode & syscall.S_IFSOCK)
	} else {
		return false
	}
}

func files() []*SocketFile {
	defer os.Unsetenv("LISTEN_PID")
	defer os.Unsetenv("LISTEN_FDS")
	defer os.Unsetenv("LISTEN_FDNAMES")

	caddy.Log().Debug("Passed File Descriptors",
		zap.String("LISTEN_PID", os.Getenv("LISTEN_PID")),
		zap.String("LISTEN_FDS", os.Getenv("LISTEN_FDS")),
		zap.String("LISTEN_FDNAMES", os.Getenv("LISTEN_FDNAMES")))

	pid, err := strconv.Atoi(os.Getenv("LISTEN_PID"))
	if err != nil || pid != os.Getpid() {
		return make([]*SocketFile, 0)
	}

	nfds, err := strconv.Atoi(os.Getenv("LISTEN_FDS"))
	if err != nil || nfds == 0 {
		return make([]*SocketFile, 0)
	}

	names := strings.Split(os.Getenv("LISTEN_FDNAMES"), ":")

	fds := make([]*SocketFile, 0, nfds)

	for i := range nfds {
		fd := SD_LISTEN_FDS_START + i
		syscall.CloseOnExec(fd)

		if !isSocket(fd) {
			continue
		}

		sotype, err := syscall.GetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_TYPE)

		if err != nil {
			continue
		}

		var name string
		if i <= len(names) {
			name = names[i]
		} else {
			name = fmt.Sprintf("%d", fd)
		}

		f := os.NewFile(uintptr(fd), name)

		if f != nil {
			fds = append(fds, &SocketFile{0, sotype, f})
		}
	}

	return fds
}

func listener(f *SocketFile) (any, error) {
	switch f.sotype {
	case syscall.SOCK_STREAM:
		return net.FileListener(f.file)
	case syscall.SOCK_DGRAM:
		return net.FilePacketConn(f.file)
	default:
		return nil, fmt.Errorf("unknown socket type %q for fd %q", f.sotype, f.file.Fd())
	}
}

func listen(addr string, predicate SocketFilePredicate) (any, error) {
	host, _, err := net.SplitHostPort(addr)

	if err == nil {
		addr = host
	}

	fd, err := strconv.Atoi(addr)
	if err == nil && fd > SD_LISTEN_FDS_START {
		i := fd - SD_LISTEN_FDS_START
		if i < len(fs) {
			f := fs[i]
			if predicate(f) {
				return listener(f)
			}
		}
	}

	for _, f := range fs {
		if f.file.Name() == addr && predicate(f) {
			return listener(f)
		}
	}

	return nil, fmt.Errorf("no matching file descriptor found for %q", addr)
}

func listenStream(ctx context.Context, network string, addr string, cfg net.ListenConfig) (any, error) {
	return listen(addr, func(f *SocketFile) bool { return f.sotype == syscall.SOCK_STREAM })
}

func listenDGRAM(ctx context.Context, network string, addr string, cfg net.ListenConfig) (any, error) {
	return listen(addr, func(f *SocketFile) bool { return f.sotype == syscall.SOCK_DGRAM })
}
