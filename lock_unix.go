//go:build unix

package compliance

import (
	"os"
	"syscall"
)

// flockExclusive takes a blocking exclusive advisory lock.
//
//nolint:wrapcheck // syscall pass-through; acquireFileLock callers wrap
func flockExclusive(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
}

// flockRelease drops the advisory lock.
//
//nolint:wrapcheck // syscall pass-through; release callers swallow or log
func flockRelease(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
