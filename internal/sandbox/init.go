package sandbox

import (
	"log"
	"sync"
)

var (
	defaultSandbox Sandbox
	sandboxOnce    sync.Once
)

// Init 初始化默认沙箱（仅执行一次，线程安全）
func Init() Sandbox {
	sandboxOnce.Do(func() {
		s, err := NewDockerSandbox()
		if err != nil {
			log.Printf("Warning: Docker sandbox unavailable (%v), falling back to host execution", err)
			defaultSandbox = nil
		} else {
			defaultSandbox = s
		}
	})
	return defaultSandbox
}

// Get 返回已初始化的沙箱实例（可能为 nil）
func Get() Sandbox {
	if defaultSandbox == nil {
		return Init()
	}
	return defaultSandbox
}
