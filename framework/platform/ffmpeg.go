package platform

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/wieku/danser-go/framework/files"
)

const errMsg = "ffmpeg not found! Please make sure it's installed in danser directory or in PATH. Follow download instructions at https://github.com/Wieku/danser-go/wiki/FFmpeg"

var ffmpegInit bool
var ffPath string

func PrepareFFMpeg(cmdName string, args ...string) (*exec.Cmd, error) {
	if !ffmpegInit {
		ffmpegInit = true

		ffmpegExec, err := files.GetCommandExec("ffmpeg", "ffmpeg")
		if err != nil {
			return nil, fmt.Errorf(errMsg)
		}

		ffPath = filepath.Dir(ffmpegExec)
		log.Println("FFmpeg exec location:", ffmpegExec)
	} else if ffPath == "" {
		return nil, fmt.Errorf(errMsg)
	}

	execPath := filepath.Join(ffPath, cmdName)

	if runtime.GOOS != "windows" {
		if stat, err := os.Stat(execPath); err == nil {
			os.Chmod(execPath, (stat.Mode()&os.ModePerm)|0111) // Just try
		}
	}

	cmd := exec.Command(execPath, args...)
	cmd.Dir = ffPath

	return cmd, nil
}
