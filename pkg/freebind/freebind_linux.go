//go:build linux
// +build linux

package freebind

import "syscall"

// ControlFreeBind is a function that can be passed to net.Dialer.Control
func ControlFreeBind(network, address string, c syscall.RawConn) error {
	if err := freeBind(network, address, c); err != nil {
		return err
	}

	return nil
}

// from https://github.com/zrepl/zrepl/blob/master/util/tcpsock/tcpsock_freebind_linux.go
func freeBind(network, address string, c syscall.RawConn) error {
	var err, sockErr error
	err = c.Control(func(fd uintptr) {
		// apparently, this works for both IPv4 and IPv6
		sockErr = syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_FREEBIND, 1)
	})
	if err != nil {
		return err
	}
	return sockErr
}
