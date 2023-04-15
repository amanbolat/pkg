package pkgnet

import (
	"errors"
	"net"
)

// RandomTCPPort returns random TCP port that is ready to use.
func RandomTCPPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer func() {
		err = l.Close()
	}()

	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.New("failed cast net.Listener.Addr() to *net.TCPAddr")
	}

	return addr.Port, nil
}
