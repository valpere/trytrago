// domain/version.go
package domain

// Version information will be injected during build using ldflags
var (
    // Version represents the current version of trytrago
    Version = "dev"
    
    // CommitSHA represents the Git commit hash used to build the program
    CommitSHA = "none"
    
    // BuildTime represents when the program was built
    BuildTime = "unknown"
    
    // BuildBy represents who or what system built the program
    BuildBy = "unknown"
)

// VersionInfo encapsulates all version-related information
type VersionInfo struct {
    Version   string
    CommitSHA string
    BuildTime string
    BuildBy   string
}

// GetVersionInfo returns the current version information
func GetVersionInfo() VersionInfo {
    return VersionInfo{
        Version:   Version,
        CommitSHA: CommitSHA,
        BuildTime: BuildTime,
        BuildBy:   BuildBy,
    }
}
