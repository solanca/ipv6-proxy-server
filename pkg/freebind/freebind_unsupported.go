//go:build !linux && !freebsd
// +build !linux,!freebsd

package freebind

import "syscall"

// ControlFreeBind leave nil
var ControlFreeBind func(network, address string, c syscall.RawConn) error = nil
