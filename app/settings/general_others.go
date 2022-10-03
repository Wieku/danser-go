//go:build !windows

package settings

func getOsuInstallation() string {
	dir, _ := os.UserHomeDir()
	return filepath.Join(dir, ".osu")
}
