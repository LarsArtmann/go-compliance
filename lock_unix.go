//go:build unix

package compliance

import (
	"os"
	"syscall"
)

// flockExclusive takes a blocking exclusive advisory lock.
func flockExclusive(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
}

// flockRelease drops the advisory lock.
func flockRelease(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
