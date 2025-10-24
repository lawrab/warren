// Package version provides version information for Warren.
package version

import (
	"fmt"
	"runtime"
)

const (
	// Major version number
	Major = 0

	// Minor version number
	Minor = 3

	// Patch version number
	Patch = 0

	// PreRelease is for pre-release versions (e.g., "alpha", "beta", "rc.1")
	// Empty string for stable releases
	PreRelease = ""
)

// Version returns the full version string
func Version() string {
	v := fmt.Sprintf("%d.%d.%d", Major, Minor, Patch)
	if PreRelease != "" {
		v += "-" + PreRelease
	}
	return v
}

// FullVersion returns version with additional build information
func FullVersion() string {
	return fmt.Sprintf("Warren v%s (%s/%s)", Version(), runtime.GOOS, runtime.GOARCH)
}

// Short returns just the version number (e.g., "v0.1.0")
func Short() string {
	return "v" + Version()
}
