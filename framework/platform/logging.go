package platform

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/wieku/danser-go/build"
	"github.com/wieku/danser-go/framework/env"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func StartLogging(logName string) {
	log.Println(build.ProgramName, "version:", build.VERSION)

	file, err := os.Create(filepath.Join(env.DataDir(), logName+".log"))
	if err != nil {
		panic(err)
	}

	log.SetOutput(file)

	PrintPlatformInfo()

	log.SetOutput(io.MultiWriter(os.Stdout, file))
}

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
