package platform

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/wieku/danser-go/build"
	"log"
	"os"
	"strings"
)

func PrintPlatformInfo() {
	osName, cpuName, ramAmount := "Unknown", "Unknown", "Unknown"

	if hStat, err := host.Info(); err == nil {
		osName = hStat.Platform + " " + hStat.PlatformVersion
	}

	if cStats, err := cpu.Info(); err == nil && len(cStats) > 0 {
		cpuName = fmt.Sprintf("%s, %d cores", strings.TrimSpace(cStats[0].ModelName), cStats[0].Cores)
	}

	if mStat, err := mem.VirtualMemory(); err == nil {
		ramAmount = humanize.IBytes(mStat.Total)
	}

	log.Println("-------------------------------------------------------------------")
	log.Println(build.ProgramName, "version:", build.VERSION)
	log.Println("Build commit hash:", build.CommitHash)
	log.Println("Ran using:", os.Args)
	log.Println("OS: ", osName)
	log.Println("CPU:", cpuName)
	log.Println("RAM:", ramAmount)
	log.Println("-------------------------------------------------------------------")
}
