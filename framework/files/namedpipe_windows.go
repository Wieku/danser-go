// +build windows

package files

import (
	"github.com/Microsoft/go-winio"
	"github.com/wieku/danser-go/framework/util"
	"net"
	"strings"
)

type NamedPipe struct {
	listener   net.Listener
	connection net.Conn

	name string
}

// NewNamedPipe creates a new named pipe to use in IPC (Inter-process communication)
// If name is empty, it generates a random 32 character hex string
// Warning: name provided is only a hint, if you want to use this pipe in IPC, retrieve the name by (*NamedPipe).Name()
func NewNamedPipe(name string) (*NamedPipe, error) {
	if strings.TrimSpace(name) == "" {
		name = util.RandomHexString(32)
	}

	name = "\\\\.\\pipe\\ro2d" + name

	listener, err := winio.ListenPipe(name, &winio.PipeConfig{
		SecurityDescriptor: "",
		MessageMode:        false,
		InputBufferSize:    0,
		OutputBufferSize:   0,
	})

	if err != nil {
		return nil, err
	}

	return &NamedPipe{
		listener: listener,
		name:     name,
	}, nil
}

func (namedPipe *NamedPipe) Read(p []byte) (n int, err error) {
	if namedPipe.connection == nil {
		namedPipe.connection, err = namedPipe.listener.Accept()

		if err != nil {
			return
		}
	}

	return namedPipe.connection.Read(p)
}

func (namedPipe *NamedPipe) Write(p []byte) (n int, err error) {
	if namedPipe.connection == nil {
		namedPipe.connection, err = namedPipe.listener.Accept()

		if err != nil {
			return
		}
	}

	return namedPipe.connection.Write(p)
}

func (namedPipe *NamedPipe) Close() (err error) {
	if namedPipe.connection != nil {
		err = namedPipe.connection.Close()
	}

	return
}

// Name returns a system name of the pipe to use in IPC
func (namedPipe *NamedPipe) Name() string {
	return namedPipe.name
}
