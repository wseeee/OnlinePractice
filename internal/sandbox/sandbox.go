package sandbox

import "context"

// Result 表示一次沙箱执行的结果
type Result struct {
	ExitCode int    // 进程退出码
	Stdout   string // 标准输出
	Stderr   string // 标准错误
	TimedOut bool   // 是否超时
	Error    error  // 沙箱内部错误（非用户代码错误）
}

// Profile 定义一种语言的编译和运行参数
type Profile struct {
	Name        string   // 语言名称，如 "go", "python"
	Ext         string   // 源文件扩展名，如 ".go", ".py"
	Image       string   // Docker 镜像
	CompileCmd  []string // 编译命令，nil 表示不需要编译
	RunCmd      []string // 运行命令，{file} 会被替换为源文件名
}

// Sandbox 沙箱执行接口
type Sandbox interface {
	// Run 编译并执行用户代码，stdin 通过 input.txt 传入
	Run(ctx context.Context, code string, stdinInput string, profile Profile, cpuCores float64, memoryMB int, timeLimitSec int) *Result
	// IsAvailable 检查沙箱是否可用
	IsAvailable(ctx context.Context) bool
}
