package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

var Client *client.Client
var Engine string

func Init(engine string, socketPath string) error {
	Engine = engine
	if Engine == "" {
		Engine = "docker"
	}

	if !isInstalled(Engine) {
		var err error
		if Engine == "podman" {
			err = installPodman()
		} else {
			err = installDocker()
		}
		if err != nil {
			return fmt.Errorf("failed to install %s: %w", Engine, err)
		}
	}

	opts := []client.Opt{client.FromEnv, client.WithAPIVersionNegotiation()}
	if socketPath != "" {
		opts = append(opts, client.WithHost("unix://"+socketPath))
	} else if Engine == "podman" {
		foundSocket := ""
		if os.Geteuid() != 0 {
			if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
				rootlessSocket := filepath.Join(runtimeDir, "podman/podman.sock")
				if _, err := os.Stat(rootlessSocket); err == nil {
					foundSocket = rootlessSocket
				}
			}
		}

		if foundSocket == "" {
			if _, err := os.Stat("/run/podman/podman.sock"); err == nil {
				foundSocket = "/run/podman/podman.sock"
			}
		}

		if foundSocket != "" {
			opts = append(opts, client.WithHost("unix://"+foundSocket))
		}
	}

	var err error
	Client, err = client.NewClientWithOpts(opts...)
	if err != nil {
		return fmt.Errorf("failed to create %s client: %w", Engine, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = Client.Ping(ctx)
	if err != nil {
		if isPermissionError(err) {
			return formatPermissionError()
		}
		return fmt.Errorf("%s daemon not responding: %w", Engine, err)
	}

	return nil
}

func isInstalled(engine string) bool {
	_, err := exec.LookPath(engine)
	return err == nil
}

func isPermissionError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "permission denied") || strings.Contains(msg, "connect: permission denied")
}

func formatPermissionError() error {
	username := "your-user"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}

	if Engine == "podman" {
		return fmt.Errorf(
			"cannot connect to the Podman socket: permission denied\n\n" +
				"Ensure the podman socket is running:\n\n" +
				"  sudo systemctl enable --now podman.socket\n\n" +
				"If you want to run rootless, ensure your socket is available at $XDG_RUNTIME_DIR/podman/podman.sock",
		)
	}

	return fmt.Errorf(
		"cannot connect to the Docker daemon: permission denied\n\n" +
			"Axis does not need to run as root. To fix this, add your user to the docker group:\n\n" +
			"  sudo usermod -aG docker " + username + "\n\n" +
			"Then log out and back in (or run: newgrp docker) for the change to take effect.\n\n" +
			"Alternatively, if you're using rootless Docker, set docker_socket in config.yaml:\n\n" +
			"  docker_socket: \"/run/user/" + fmt.Sprintf("%d", os.Getuid()) + "/docker.sock\"",
	)
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
