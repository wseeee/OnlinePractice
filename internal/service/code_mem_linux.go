//go:build linux

package service

import (
	"os"
	"syscall"
)

func getChildMaxRSS(ps *os.ProcessState) int {
	if ps == nil {
		return 0
	}
	if usage, ok := ps.SysUsage().(*syscall.Rusage); ok {
		return int(usage.Maxrss) // KB on Linux
	}
	return 0
}
