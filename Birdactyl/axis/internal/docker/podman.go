package docker

import (
	"fmt"
	"os/exec"
)

func installPodman() error {
	var cmd *exec.Cmd

	if _, err := exec.LookPath("apt-get"); err == nil {
		script := `
			apt-get update
			apt-get install -y podman
			systemctl enable --now podman.socket
		`
		cmd = exec.Command("bash", "-c", script)
	} else if _, err := exec.LookPath("dnf"); err == nil {
		script := `
			dnf install -y podman
			systemctl enable --now podman.socket
		`
		cmd = exec.Command("bash", "-c", script)
	} else if _, err := exec.LookPath("yum"); err == nil {
		script := `
			yum install -y podman
			systemctl enable --now podman.socket
		`
		cmd = exec.Command("bash", "-c", script)
	} else if _, err := exec.LookPath("pacman"); err == nil {
		script := `
			pacman -Sy --noconfirm podman
			systemctl enable --now podman.socket
		`
		cmd = exec.Command("bash", "-c", script)
	} else if _, err := exec.LookPath("zypper"); err == nil {
		script := `
			zypper install -y podman
			systemctl enable --now podman.socket
		`
		cmd = exec.Command("bash", "-c", script)
	} else {
		return fmt.Errorf("unsupported linux distribution for podman: install manually")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("podman installation failed: %s: %w", string(output), err)
	}

	return nil
}
