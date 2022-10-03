//go:build windows

package settings

import (
	"golang.org/x/sys/windows/registry"
	"os"
	"path/filepath"
	"strings"
)

func getOsuInstallation() (path string) {
	path = filepath.Join(os.Getenv("localappdata"), "osu!")

	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Classes\osu!\shell\open\command`, registry.QUERY_VALUE)
	if err != nil {
		return
	}

	defer key.Close()

	s, _, err := key.GetStringValue("")
	if err != nil {
		return
	}

	// Extracting exe path from (example): "D:\osu!\osu!.exe" "%1"
	if i := strings.IndexRune(s, '"'); i > -1 {
		s = s[i+1:]

		if i = strings.IndexRune(s, '"'); i > -1 {
			s = s[:i]
		}
	}

	path = filepath.Dir(s)

	return
}
