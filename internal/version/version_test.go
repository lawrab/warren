package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	v := Version()

	// Should follow semantic versioning format
	if v == "" {
		t.Error("Version() should not return empty string")
	}

	// Should contain version numbers
	if !strings.Contains(v, ".") {
		t.Error("Version should contain dots separating numbers")
	}

	// For current version (0.1.1), verify exact format
	expected := "0.1.1"
	if v != expected {
		t.Logf("Version format changed from %s to %s - update test if intentional", expected, v)
	}
}

func TestVersionWithPreRelease(t *testing.T) {
	// Save original value
	originalPreRelease := PreRelease

	// Test with empty pre-release (stable)
	v := Version()
	if strings.Contains(v, "-") {
		t.Error("Stable version should not contain hyphen")
	}

	// Note: We can't modify constants, so this test just verifies behavior
	// with the current PreRelease value
	if PreRelease != "" {
		if !strings.Contains(v, "-") {
			t.Error("Pre-release version should contain hyphen")
		}
		if !strings.Contains(v, PreRelease) {
			t.Errorf("Version should contain pre-release tag %s", PreRelease)
		}
	}

	// Verify we haven't accidentally modified the constant
	if PreRelease != originalPreRelease {
		t.Error("Test should not modify PreRelease constant")
	}
}

func TestShort(t *testing.T) {
	short := Short()

	// Should start with 'v'
	if !strings.HasPrefix(short, "v") {
		t.Error("Short() should return version with 'v' prefix")
	}

	// Should be in format v0.1.1
	expectedPrefix := "v0.1."
	if !strings.HasPrefix(short, expectedPrefix) {
		t.Logf("Short version format changed - expected prefix %s, got %s", expectedPrefix, short)
	}

	// Should equal "v" + Version()
	expected := "v" + Version()
	if short != expected {
		t.Errorf("Short() = %q, want %q", short, expected)
	}
}

func TestFullVersion(t *testing.T) {
	full := FullVersion()

	// Should contain "Warren"
	if !strings.Contains(full, "Warren") {
		t.Error("FullVersion should contain 'Warren'")
	}

	// Should contain the version
	if !strings.Contains(full, Version()) {
		t.Errorf("FullVersion should contain version %s", Version())
	}

	// Should contain OS
	if !strings.Contains(full, runtime.GOOS) {
		t.Errorf("FullVersion should contain OS %s", runtime.GOOS)
	}

	// Should contain architecture
	if !strings.Contains(full, runtime.GOARCH) {
		t.Errorf("FullVersion should contain architecture %s", runtime.GOARCH)
	}

	// Should be in format "Warren v0.1.1 (linux/amd64)" or similar
	if !strings.Contains(full, "(") || !strings.Contains(full, ")") {
		t.Error("FullVersion should contain parentheses around OS/arch")
	}

	if !strings.Contains(full, "/") {
		t.Error("FullVersion should contain '/' separating OS and architecture")
	}
}

func TestVersionConstants(t *testing.T) {
	// Verify version constants are valid
	if Major < 0 {
		t.Error("Major version should be non-negative")
	}

	if Minor < 0 {
		t.Error("Minor version should be non-negative")
	}

	if Patch < 0 {
		t.Error("Patch version should be non-negative")
	}

	// Current version should be 0.1.1
	if Major != 0 {
		t.Logf("Major version changed from 0 to %d", Major)
	}

	if Minor != 1 {
		t.Logf("Minor version changed from 1 to %d", Minor)
	}

	if Patch != 1 {
		t.Logf("Patch version changed from 1 to %d", Patch)
	}

	// PreRelease should be a string (possibly empty for stable)
	// Just verify it's accessible and doesn't cause panic
	_ = PreRelease
}

func TestVersionConsistency(t *testing.T) {
	// All version functions should use the same underlying version
	v := Version()
	short := Short()
	full := FullVersion()

	// Short should contain the same version as Version()
	if !strings.Contains(short, v) {
		t.Errorf("Short() %q should contain Version() %q", short, v)
	}

	// Full should contain the same version as Version()
	if !strings.Contains(full, v) {
		t.Errorf("FullVersion() %q should contain Version() %q", full, v)
	}
}
