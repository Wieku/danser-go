package build

import (
	"github.com/wieku/danser-go/framework/math/mutils"
	"runtime/debug"
)

var ProgramName = "danser-go"

var CommitHash = "Unknown"

var VERSION = "dev"

var Stream = "Dev"

var DanserExec = "danser-cli"

func init() {
	if bI, ok := debug.ReadBuildInfo(); ok {
		for _, k := range bI.Settings {
			if k.Key == "vcs.revision" {
				CommitHash = k.Value

				break
			}
		}
	}

	if VERSION == "dev" {
		VERSION += "-" + CommitHash[:mutils.Min(7, len(CommitHash))]
	}
}
