package sandbox

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type DockerSandbox struct {
	cli    *client.Client
	useCLI bool // Windows Docker Desktop fallback
}

// NewDockerSandbox 创建 Docker 沙箱
func NewDockerSandbox() (*DockerSandbox, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	s := &DockerSandbox{cli: cli}
	if !s.IsAvailable(context.Background()) {
		return nil, fmt.Errorf("docker daemon not available")
	}
	log.Println("Docker sandbox initialized")
	return s, nil
}

func (s *DockerSandbox) IsAvailable(ctx context.Context) bool {
	if _, err := s.cli.Ping(ctx); err == nil {
		return true
	}
	// fallback: 某些 Windows Docker Desktop 配置下 SDK 连接可能失败，尝试 CLI
	if err := exec.Command("docker", "version").Run(); err == nil {
		cli2, err2 := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err2 == nil {
			s.cli = cli2
		}
		return true
	}
	return false
}

func (s *DockerSandbox) Run(ctx context.Context, code string, stdinInput string, profile Profile, cpuCores float64, memoryMB int, timeLimitSec int) *Result {
	ext := profile.Ext
	if ext == "" {
		ext = ".txt"
	}
	filename := "main" + ext

	// 镜像拉取
	if err := s.pullImage(ctx, profile.Image); err != nil {
		return &Result{Error: fmt.Errorf("image pull: %w", err)}
	}

	// 编译（如果需要）
	if len(profile.CompileCmd) > 0 {
		compileResult := s.runContainer(ctx, profile.Image, filename, code, stdinInput, profile.CompileCmd, cpuCores, float64(memoryMB), min(timeLimitSec, 30))
		if compileResult.Error != nil || compileResult.ExitCode != 0 {
			return compileResult
		}
	}

	// 运行
	return s.runContainer(ctx, profile.Image, filename, code, stdinInput, profile.RunCmd, cpuCores, float64(memoryMB), timeLimitSec)
}

func (s *DockerSandbox) runContainer(ctx context.Context, imageName, filename, code, stdinInput string, cmd []string, cpuCores, memoryMB float64, timeLimitSec int) *Result {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeLimitSec)*time.Second)
	defer cancel()

	codeTar := buildCodeTar(filename, code, stdinInput)
	memBytes := int64(memoryMB * 1024 * 1024)
	nanoCPUs := int64(cpuCores * 1e9)

	contResp, err := s.cli.ContainerCreate(timeoutCtx,
		&container.Config{
			Image:      imageName,
			Cmd:        cmd,
			WorkingDir: "/code",
		},
		&container.HostConfig{
			AutoRemove:  false,
			NetworkMode: "none",
			Resources: container.Resources{
				Memory:   memBytes,
				NanoCPUs: nanoCPUs,
			},
			CapDrop:     []string{"ALL"},
			SecurityOpt: []string{"no-new-privileges:true"},
		},
		nil, nil, "",
	)
	if err != nil {
		return &Result{Error: fmt.Errorf("container create: %w", err)}
	}
	containerID := contResp.ID
	defer s.cleanup(containerID)

	if err := s.cli.CopyToContainer(timeoutCtx, containerID, "/code", codeTar, container.CopyToContainerOptions{}); err != nil {
		return &Result{Error: fmt.Errorf("copy code: %w", err)}
	}

	if err := s.cli.ContainerStart(timeoutCtx, containerID, container.StartOptions{}); err != nil {
		return &Result{Error: fmt.Errorf("container start: %w", err)}
	}

	statusCh, errCh := s.cli.ContainerWait(timeoutCtx, containerID, container.WaitConditionNotRunning)
	var exitCode int
	timedOut := false

	select {
	case status := <-statusCh:
		exitCode = int(status.StatusCode)
	case err := <-errCh:
		if timeoutCtx.Err() != nil {
			timedOut = true
			s.cli.ContainerKill(ctx, containerID, "KILL")
			<-statusCh
		} else {
			return &Result{Error: fmt.Errorf("container wait: %w", err)}
		}
	case <-timeoutCtx.Done():
		timedOut = true
		s.cli.ContainerKill(ctx, containerID, "KILL")
		select {
		case <-statusCh:
		case <-errCh:
		}
	}

	logsReader, err := s.cli.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return &Result{Error: fmt.Errorf("container logs: %w", err)}
	}
	defer logsReader.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	stdcopy.StdCopy(&stdoutBuf, &stderrBuf, logsReader)

	return &Result{
		ExitCode: exitCode,
		Stdout:   strings.TrimSpace(stdoutBuf.String()),
		Stderr:   truncate(stderrBuf.String(), 500),
		TimedOut: timedOut,
	}
}

func (s *DockerSandbox) pullImage(ctx context.Context, imageName string) error {
	_, _, err := s.cli.ImageInspectWithRaw(ctx, imageName)
	if err == nil {
		return nil
	}

	log.Printf("Pulling Docker image: %s", imageName)
	reader, err := s.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	io.Copy(io.Discard, reader)
	return nil
}

func (s *DockerSandbox) cleanup(containerID string) {
	if err := s.cli.ContainerRemove(context.Background(), containerID, container.RemoveOptions{Force: true}); err != nil {
		log.Printf("[Sandbox] cleanup %s: %v", containerID[:12], err)
	}
}

func buildCodeTar(filename, code, stdinInput string) io.Reader {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	tw.WriteHeader(&tar.Header{
		Name: filename,
		Size: int64(len(code)),
		Mode: 0644,
	})
	tw.Write([]byte(code))

	if stdinInput != "" {
		tw.WriteHeader(&tar.Header{
			Name: "input.txt",
			Size: int64(len(stdinInput)),
			Mode: 0644,
		})
		tw.Write([]byte(stdinInput))
	}

	tw.Close()
	return bytes.NewReader(buf.Bytes())
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
