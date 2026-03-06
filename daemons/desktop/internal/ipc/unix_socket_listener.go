//go:builld linux

package ipc

import (
	"net"
)

const SocketFileName = "/temp/watcher.sock"

func IPCListener() (net.Listener, error) {
	return net.Listen("unix", SocketFileName)
}
