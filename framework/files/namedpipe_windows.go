//go:build windows

package files

import (
	"github.com/Microsoft/go-winio"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/util"
	"net"
	"strings"
	"sync"
)

type NamedPipe struct {
	listener   net.Listener
	connection net.Conn

	wg      *sync.WaitGroup
	connErr error

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

	listener, err := winio.ListenPipe(name, nil)

	if err != nil {
		return nil, err
	}

	pipe := &NamedPipe{
		listener: listener,
		wg:       &sync.WaitGroup{},
		name:     name,
	}

	pipe.wg.Add(1)

	wg2 := &sync.WaitGroup{}
	wg2.Add(1)

	goroutines.Run(func() {
		wg2.Done()

		pipe.connection, pipe.connErr = pipe.listener.Accept()

		pipe.wg.Done()
	})

	// wait until goroutine starts
	wg2.Wait()

	return pipe, nil
}

func (namedPipe *NamedPipe) Read(p []byte) (n int, err error) {
	namedPipe.wg.Wait()

	if namedPipe.connErr != nil {
		return -1, namedPipe.connErr
	}

	return namedPipe.connection.Read(p)
}

func (namedPipe *NamedPipe) Write(p []byte) (n int, err error) {
	namedPipe.wg.Wait()

	if namedPipe.connErr != nil {
		return -1, namedPipe.connErr
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
