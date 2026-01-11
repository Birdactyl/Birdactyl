package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

var Client *client.Client

func Init() error {
	if !IsInstalled() {
		if err := Install(); err != nil {
			return fmt.Errorf("failed to install docker: %w", err)
		}
	}

	var err error
	Client, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = Client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("docker daemon not responding: %w", err)
	}

	return nil
}

func IsInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func Install() error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("automatic docker installation only supported on linux")
	}

	var cmd *exec.Cmd

	if _, err := exec.LookPath("apt-get"); err == nil {
		script := `
			apt-get update
			apt-get install -y ca-certificates curl gnupg
			install -m 0755 -d /etc/apt/keyrings
			curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
			chmod a+r /etc/apt/keyrings/docker.gpg
			echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
			apt-get update
			apt-get install -y docker-ce docker-ce-cli containerd.io
			systemctl enable docker
			systemctl start docker
		`
		cmd = exec.Command("bash", "-c", script)
	} else if _, err := exec.LookPath("dnf"); err == nil {
		script := `
			dnf -y install dnf-plugins-core
			dnf config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo
			dnf install -y docker-ce docker-ce-cli containerd.io
			systemctl enable docker
			systemctl start docker
		`
		cmd = exec.Command("bash", "-c", script)
	} else if _, err := exec.LookPath("yum"); err == nil {
		script := `
			yum install -y yum-utils
			yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
			yum install -y docker-ce docker-ce-cli containerd.io
			systemctl enable docker
			systemctl start docker
		`
		cmd = exec.Command("bash", "-c", script)
	} else {
		return fmt.Errorf("unsupported linux distribution")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("installation failed: %s: %w", string(output), err)
	}

	return nil
}

func PullImage(ctx context.Context, imageName string) error {
	reader, err := Client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	io.Copy(io.Discard, reader)
	return nil
}

func ImageExists(ctx context.Context, imageName string) bool {
	_, _, err := Client.ImageInspectWithRaw(ctx, imageName)
	return err == nil
}

func CreateContainer(ctx context.Context, name string, cfg *container.Config, hostCfg *container.HostConfig) (string, error) {
	resp, err := Client.ContainerCreate(ctx, cfg, hostCfg, nil, nil, name)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func StartContainer(ctx context.Context, id string) error {
	return Client.ContainerStart(ctx, id, container.StartOptions{})
}

func StopContainer(ctx context.Context, id string, timeout int) error {
	t := timeout
	return Client.ContainerStop(ctx, id, container.StopOptions{Timeout: &t})
}

func KillContainer(ctx context.Context, id string) error {
	return Client.ContainerKill(ctx, id, "SIGKILL")
}

func RestartContainer(ctx context.Context, id string, timeout int) error {
	t := timeout
	return Client.ContainerRestart(ctx, id, container.StopOptions{Timeout: &t})
}

func RemoveContainer(ctx context.Context, id string, force bool) error {
	return Client.ContainerRemove(ctx, id, container.RemoveOptions{Force: force})
}

func ContainerExists(ctx context.Context, name string) bool {
	name = strings.TrimPrefix(name, "/")
	containers, err := Client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return false
	}
	for _, c := range containers {
		for _, n := range c.Names {
			if strings.TrimPrefix(n, "/") == name {
				return true
			}
		}
	}
	return false
}

func GetContainerID(ctx context.Context, name string) (string, error) {
	name = strings.TrimPrefix(name, "/")
	containers, err := Client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return "", err
	}
	for _, c := range containers {
		for _, n := range c.Names {
			if strings.TrimPrefix(n, "/") == name {
				return c.ID, nil
			}
		}
	}
	return "", fmt.Errorf("container not found: %s", name)
}


func GetContainerLogs(ctx context.Context, id string, tail string, follow bool) (io.ReadCloser, error) {
	return Client.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Tail:       tail,
		Timestamps: true,
	})
}


func GetContainerStats(ctx context.Context, id string) (*types.StatsJSON, error) {
	resp, err := Client.ContainerStats(ctx, id, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stats types.StatsJSON
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

func SendCommand(ctx context.Context, id string, command string) error {
	hijacked, err := Client.ContainerAttach(ctx, id, container.AttachOptions{
		Stdin:  true,
		Stream: true,
	})
	if err != nil {
		return err
	}
	defer hijacked.Close()

	_, err = hijacked.Conn.Write([]byte(command + "\n"))
	if err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)
	return nil
}
