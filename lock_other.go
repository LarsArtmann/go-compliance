//go:build !unix

package compliance

import "os"

// flockExclusive is unavailable off unix; reporters fall back to
// last-writer-wins semantics rather than pretending to lock.
func flockExclusive(*os.File) error { return nil }

// flockRelease is unavailable off unix.
func flockRelease(*os.File) error { return nil }
