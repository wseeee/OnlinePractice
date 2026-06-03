//go:build !linux

package service

import "os"

// getChildMaxRSS 获取子进程 Max RSS（KB），非 Linux 平台暂不支持，返回 0 跳过内存检查。
func getChildMaxRSS(ps *os.ProcessState) int {
	return 0
}
