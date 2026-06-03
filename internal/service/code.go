package service

import (
	"OnlinePrictice/internal/helper"
	"OnlinePrictice/internal/model"
	"OnlinePrictice/internal/sandbox"
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const maxOutputSize = 1024 * 1024 // 1MB 输出限制

// SaveAndValidateCodeLang 保存代码 + 按语言校验
func SaveAndValidateCodeLang(code []byte, language string) (codePath string, valid bool, err error) {
	path, err := helper.SaveCodeTo(code, language)
	if err != nil {
		return "", false, err
	}
	// 非 Go 语言跳过 import 校验
	if language != "go" {
		return path, true, nil
	}
	v, err := helper.CheckGoCodeValid(path)
	if err != nil {
		return "", false, err
	}
	return path, v, nil
}

// JudgeCodeWithLang 判题（支持多语言）
func JudgeCodeWithLang(codePath string, problemIdentity string, language string) (passed, total int, status int, err error) {
	pb := new(model.ProblemBasic)
	err = model.DB.Where("identity = ?", problemIdentity).Preload("TestCases").First(pb).Error
	if err != nil {
		return 0, 0, 0, err
	}
	if len(pb.TestCases) == 0 {
		return 0, 0, 0, errors.New("该题目没有测试用例")
	}
	total = len(pb.TestCases)

	// 读取代码
	codeBytes, err := os.ReadFile(codePath)
	if err != nil {
		return 0, total, 0, err
	}
	code := string(codeBytes)

	// Go 代码防御性校验
	if language == "go" {
		v, err := helper.CheckGoCodeValid(codePath)
		if err != nil {
			return 0, total, 0, err
		}
		if !v {
			return 0, total, 6, nil
		}
	}

	// 获取语言配置
	profile, ok := sandbox.GetProfile(language)
	if !ok {
		return 0, total, 0, errors.New("不支持的语言: " + language)
	}

	// 获取沙箱
	sb := sandbox.Get()

	if sb != nil {
		// 沙箱路径
		return judgeInSandbox(sb, code, profile, pb)
	}
	// 降级：本地 exec 执行
	switch language {
	case "go":
		return judgeLocalGo(codePath, pb)
	case "python":
		return judgeLocalPython(codePath, pb)
	case "cpp", "c++":
		return judgeLocalCpp(codePath, pb)
	}
	return 0, total, 0, errors.New("沙箱不可用，当前语言不支持本地执行: "+language)
}

// judgeInSandbox Docker 沙箱判题
func judgeInSandbox(sb sandbox.Sandbox, code string, profile sandbox.Profile, pb *model.ProblemBasic) (passed, total int, status int, err error) {
	total = len(pb.TestCases)
	ctx := context.Background()

	passCount := 0
	var lock sync.Mutex

	WA := make(chan int, total)
	CE := make(chan string, total)
	AC := make(chan int, total)
	OOM := make(chan int)

	cpuCores := 0.5
	memoryMB := pb.MaxMem
	if memoryMB <= 0 {
		memoryMB = 256
	}
	timeLimitSec := pb.MaxRuntime / 1000
	if timeLimitSec <= 0 {
		timeLimitSec = 5
	}

	var wg sync.WaitGroup
	for _, tc := range pb.TestCases {
		tc := tc
		wg.Add(1)
		go func() {
			defer wg.Done()

			result := sb.Run(ctx, code, tc.Input, profile, cpuCores, memoryMB, timeLimitSec)
			if result.Error != nil {
				log.Printf("[Sandbox] error: %v", result.Error)
				CE <- "内部错误"
				return
			}
			if result.TimedOut {
				OOM <- 1
				return
			}
			if result.ExitCode != 0 && result.Stderr != "" {
				CE <- result.Stderr
				return
			}

			actual := strings.TrimSpace(result.Stdout)
			expected := strings.TrimSpace(tc.Output)
			if expected != actual {
				WA <- 1
				return
			}

			lock.Lock()
			passCount++
			if passCount == total {
				AC <- 1
			}
			lock.Unlock()
		}()
	}

	go func() {
		wg.Wait()
		close(WA)
		close(CE)
		close(AC)
	}()

	select {
	case <-WA:
		lock.Lock()
		return passCount, total, 2, nil
	case <-OOM:
		lock.Lock()
		return passCount, total, 3, nil // TLE
	case errMsg := <-CE:
		lock.Lock()
		_ = errMsg
		return passCount, total, 5, nil
	case <-AC:
		lock.Lock()
		return total, total, 1, nil
	case <-time.After(time.Millisecond*time.Duration(pb.MaxRuntime) + 3*time.Second):
		lock.Lock()
		if passCount == total {
			return total, total, 1, nil
		}
		return passCount, total, 3, nil
	}
}

// judgeLocalGo 本地 exec 执行 Go 代码（降级方案）
// 先 go build 编译一次，再对每个测试用例运行编译好的二进制，避免重复编译导致超时
func judgeLocalGo(codePath string, pb *model.ProblemBasic) (passed, total int, status int, err error) {
	total = len(pb.TestCases)

	v, err := helper.CheckGoCodeValid(codePath)
	if err != nil {
		return 0, total, 0, err
	}
	if !v {
		return 0, total, 6, nil
	}

	// 先编译一次
	binPath := codePath + ".exe"
	buildCmd := exec.Command("go", "build", "-o", binPath, codePath)
	var buildStderr bytes.Buffer
	buildCmd.Stderr = &buildStderr
	if err := buildCmd.Run(); err != nil {
		return 0, total, 5, nil // CE
	}
	defer os.Remove(binPath)

	WA := make(chan int, total)
	OOM := make(chan int, total)
	CE := make(chan string, total)
	AC := make(chan int, total)

	passCount := 0
	var lock sync.Mutex

	var wg sync.WaitGroup
	for _, testCase := range pb.TestCases {
		testCase := testCase
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.Command(binPath)
			var out, stderr bytes.Buffer
			cmd.Stderr = &stderr
			cmd.Stdout = &out
			stdinPipe, err := cmd.StdinPipe()
			if err != nil {
				CE <- "内部错误"
				return
			}
			io.WriteString(stdinPipe, testCase.Input+"\n")
			stdinPipe.Close()

			if err := cmd.Run(); err != nil {
				// 运行时错误（非零退出码），视为 WA
			}

			actual := strings.TrimSpace(out.String())
			expected := strings.TrimSpace(testCase.Output)
			if expected != actual {
				WA <- 1
				return
			}
			if rss := getChildMaxRSS(cmd.ProcessState); rss > 0 && rss > pb.MaxMem {
				OOM <- 1
				return
			}
			lock.Lock()
			passCount++
			if passCount == total {
				AC <- 1
			}
			lock.Unlock()
		}()
	}

	go func() {
		wg.Wait()
		close(WA)
		close(OOM)
		close(CE)
		close(AC)
	}()

	select {
	case <-WA:
		lock.Lock()
		return passCount, total, 2, nil
	case <-OOM:
		lock.Lock()
		return passCount, total, 4, nil
	case errMsg := <-CE:
		lock.Lock()
		_ = errMsg
		return passCount, total, 5, nil
	case <-AC:
		lock.Lock()
		return total, total, 1, nil
	case <-time.After(time.Millisecond * time.Duration(pb.MaxRuntime)):
		lock.Lock()
		if passCount == total {
			return total, total, 1, nil
		}
		return passCount, total, 3, nil
	}
}

// judgeLocalPython 本地执行 Python 代码（降级方案）
func judgeLocalPython(codePath string, pb *model.ProblemBasic) (passed, total int, status int, err error) {
	total = len(pb.TestCases)

	// 查找 python3 或 python
	pythonBin := "python3"
	if _, err := exec.LookPath("python3"); err != nil {
		if _, err := exec.LookPath("python"); err != nil {
			return 0, total, 0, errors.New("未找到 Python 解释器")
		}
		pythonBin = "python"
	}

	return judgeLocalRun(codePath, pb, pythonBin, codePath)
}

// judgeLocalCpp 本地执行 C++ 代码（降级方案）
func judgeLocalCpp(codePath string, pb *model.ProblemBasic) (passed, total int, status int, err error) {
	total = len(pb.TestCases)

	if _, err := exec.LookPath("g++"); err != nil {
		return 0, total, 0, errors.New("未找到 g++ 编译器")
	}

	// 编译
	binPath := codePath + ".exe"
	buildCmd := exec.Command("g++", "-O2", "-std=c++17", "-o", binPath, codePath)
	var buildStderr bytes.Buffer
	buildCmd.Stderr = &buildStderr
	if err := buildCmd.Run(); err != nil {
		return 0, total, 5, nil // CE
	}
	defer os.Remove(binPath)

	return judgeLocalRun(codePath, pb, binPath)
}

// judgeLocalRun 执行编译好的二进制/脚本，对每个测试用例并行运行
func judgeLocalRun(_ string, pb *model.ProblemBasic, binary string, args ...string) (passed, total int, status int, err error) {
	total = len(pb.TestCases)

	WA := make(chan int, total)
	OOM := make(chan int, total)
	CE := make(chan string, total)
	AC := make(chan int, total)

	passCount := 0
	var lock sync.Mutex

	var wg sync.WaitGroup
	for _, testCase := range pb.TestCases {
		testCase := testCase
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.Command(binary, args...)
			var out, stderr bytes.Buffer
			cmd.Stderr = &stderr
			cmd.Stdout = &out
			stdinPipe, err := cmd.StdinPipe()
			if err != nil {
				CE <- "内部错误"
				return
			}
			io.WriteString(stdinPipe, testCase.Input+"\n")
			stdinPipe.Close()

			cmd.Run()

			actual := strings.TrimSpace(out.String())
			expected := strings.TrimSpace(testCase.Output)
			if expected != actual {
				WA <- 1
				return
			}
			lock.Lock()
			passCount++
			if passCount == total {
				AC <- 1
			}
			lock.Unlock()
		}()
	}

	go func() {
		wg.Wait()
		close(WA)
		close(OOM)
		close(CE)
		close(AC)
	}()

	select {
	case <-WA:
		lock.Lock()
		return passCount, total, 2, nil
	case <-OOM:
		lock.Lock()
		return passCount, total, 4, nil
	case errMsg := <-CE:
		lock.Lock()
		_ = errMsg
		return passCount, total, 5, nil
	case <-AC:
		lock.Lock()
		return total, total, 1, nil
	case <-time.After(time.Millisecond * time.Duration(pb.MaxRuntime)):
		lock.Lock()
		if passCount == total {
			return total, total, 1, nil
		}
		return passCount, total, 3, nil
	}
}
