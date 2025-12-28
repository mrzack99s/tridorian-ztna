package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the current version of the application.
	Version = "0.1.0"
	// Commit is the git commit hash.
	Commit = "none"
	// BuildTime is the timestamp of the build.
	BuildTime = "unknown"
)

// Info holds version information.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// GetInfo returns the version information.
func GetInfo() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// CheckVersionCompatibility checks if the client version is compatible with the server.
// This is a placeholder for actual compatibility logic.
func CheckVersionCompatibility(clientVersion string) bool {
	// For now, assume all versions are compatible or implement semver check
	return true
}

func String() string {
	return fmt.Sprintf("Version: %s, Commit: %s, BuildTime: %s, Go: %s", Version, Commit, BuildTime, runtime.Version())
}
