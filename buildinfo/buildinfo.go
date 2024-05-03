package buildinfo

import (
	"fmt"
	"runtime"
)

// Fields injected by goreleaser.
//
//nolint:gochecknoglobals // These are injected by goreleaser.
var (
	version    = "<unknown>"
	commitDate = "date unknown"
	commit     = ""
)

// Version returns the version of the application, or "<unknown>" if it is not
// set. This is injected by goreleaser.
func Version() string {
	return version
}

// CommitDate returns the date of the commit, or "date unknown" if it is not
// set. This is injected by goreleaser.
func CommitDate() string {
	return commitDate
}

// Commit returns the commit hash, or "" if it is not set. This is injected by
// goreleaser.
func Commit() string {
	return commit
}

// Target returns the target operating system that this binary was compiled for.
func Target() string {
	return runtime.GOOS
}

// FullVersion returns the full version string, including the version, commit
// hash, and date.
func FullVersion() string {
	return fmt.Sprintf("%s %s/%s %s (%s) %s",
		version, runtime.GOOS, runtime.GOARCH, runtime.Version(), commitDate, commit)
}
