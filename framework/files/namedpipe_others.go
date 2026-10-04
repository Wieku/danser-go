//go:build !windows

package files

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/wieku/danser-go/framework/util"
)

type NamedPipe struct {
	file *os.File

	path string
	name string
}

// NewNamedPipe creates a new named pipe to use in IPC (Inter-process communication)
// If name is empty, it generates a random 32 character hex string
// Warning: name provided is only a hint, if you want to use this pipe in IPC, retrieve the name by (*NamedPipe).Name()
func NewNamedPipe(path, name string) (*NamedPipe, error) {
	if strings.TrimSpace(name) == "" {
		name = util.RandomHexString(32)
	}

	name = ".ro2d" + name

	fPath := filepath.Join(path, name)

	os.Remove(fPath)

	err := syscall.Mkfifo(fPath, 0666)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(fPath, os.O_RDWR, os.ModeNamedPipe)
	if err != nil {
		return nil, err
	}

	_, err = unix.FcntlInt(file.Fd(), unix.F_SETPIPE_SZ, 65536)
	if err != nil {
		return nil, err
	}

	return &NamedPipe{
		file: file,
		path: path,
		name: name,
	}, nil
}

func (namedPipe *NamedPipe) Read(p []byte) (n int, err error) {
	return namedPipe.file.Read(p)
}

func (namedPipe *NamedPipe) Write(p []byte) (n int, err error) {
	return namedPipe.file.Write(p)
}

func (namedPipe *NamedPipe) Close() (err error) {
	err = namedPipe.file.Close()
	if err != nil {
		return
	}

	return os.Remove(namedPipe.name)
}

// Name returns a system name of the pipe to use in IPC
func (namedPipe *NamedPipe) Name() string {
	return namedPipe.name
}

// Path returns full path of pipe file
func (namedPipe *NamedPipe) Path() string {
	return filepath.Join(namedPipe.path, namedPipe.name)
}
