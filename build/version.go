package build

import (
	"github.com/wieku/danser-go/framework/math/mutils"
	"runtime/debug"
)

var VERSION = "dev"

var Stream = "Dev"

func init() {
	if VERSION == "dev" {
		if bI, ok := debug.ReadBuildInfo(); ok {
			rev := ""

			for _, k := range bI.Settings {
				if k.Key == "vcs.revision" {
					rev = k.Value
				}
			}

			if rev != "" {
				VERSION += "-" + rev[:mutils.Min(7, len(rev))]
			}
		}
	}
}
