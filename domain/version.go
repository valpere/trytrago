// domain/version.go
package domain

import (
	"fmt"
	"runtime"
	"time"
)

// Version information will be injected during build using ldflags
var (
	// Version represents the semantic version of the application
	Version = "0.1.0"

	// CommitSHA represents the Git commit hash used to build the program
	CommitSHA = "none"

	// BuildTime represents when the program was built, in ISO 8601 format
	BuildTime = "unknown"
)

// VersionInfo encapsulates version-related information and system details
type VersionInfo struct {
	// Application version details
	Version   string    // Semantic version number
	CommitSHA string    // Git commit hash
	BuildTime time.Time // Build timestamp

	// Runtime information
	GoVersion string // Go runtime version
	GOOS      string // Operating system
	GOARCH    string // Architecture
}

// GetVersionInfo returns the current version information along with runtime details
func GetVersionInfo() (*VersionInfo, error) {
	// Parse the build time from ISO 8601 format
	buildTime, err := time.Parse(time.RFC3339, BuildTime)
	if err != nil {
		// If parsing fails, default to Unix epoch to indicate unknown build time
		buildTime = time.Unix(0, 0)
	}

	return &VersionInfo{
		Version:   Version,
		CommitSHA: CommitSHA,
		BuildTime: buildTime,
		GoVersion: runtime.Version(),
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
	}, nil
}

// String returns a formatted string representation of version information
func (v *VersionInfo) String() string {
	return fmt.Sprintf(
		"TryTraGo Dictionary Server\n"+
			"Version:    %s\n"+
			"Commit:     %s\n"+
			"Built:      %s\n"+
			"Go Version: %s\n"+
			"OS/Arch:    %s/%s",
		v.Version,
		v.CommitSHA,
		v.BuildTime.Format(time.RFC3339),
		v.GoVersion,
		v.GOOS,
		v.GOARCH,
	)
}
