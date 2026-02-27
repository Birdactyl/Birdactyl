package docker

import (
	"fmt"
	"os/exec"
)

func installDocker() error {
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
	} else if _, err := exec.LookPath("pacman"); err == nil {
		script := `
			pacman -Sy --noconfirm docker
			systemctl enable docker
			systemctl start docker
		`
		cmd = exec.Command("bash", "-c", script)
	} else if _, err := exec.LookPath("zypper"); err == nil {
		script := `
			zypper install -y docker
			systemctl enable docker
			systemctl start docker
		`
		cmd = exec.Command("bash", "-c", script)
	} else {
		return fmt.Errorf("unsupported linux distribution for docker: install manually")
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker installation failed: %s: %w", string(output), err)
	}

	return nil
}
