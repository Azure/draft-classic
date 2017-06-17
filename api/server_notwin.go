// +build !windows

package api

import (
	"net"
	"net/http"
	"os"
	"syscall"
)

func setupUnixHTTP(addr string) (*Server, error) {
	if err := syscall.Unlink(addr); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	mask := syscall.Umask(0777)
	defer syscall.Umask(mask)

	l, err := net.Listen("unix", addr)
	if err != nil {
		return nil, err
	}

	if err := os.Chmod(addr, 0660); err != nil {
		return nil, err
	}

	a := &Server{
		HTTPServer: &http.Server{Addr: addr},
		Listener:   l,
	}
	return a, nil
}
