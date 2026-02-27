package server

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CompressPath(serverID, srcPath, destPath, format string) error {
	base := serverDataDir(serverID)
	src := filepath.Join(base, filepath.Clean("/"+srcPath))
	dest := filepath.Join(base, filepath.Clean("/"+destPath))

	if !strings.HasPrefix(src, base) || !strings.HasPrefix(dest, base) {
		return fmt.Errorf("invalid path")
	}

	dest = uniquePath(dest)

	var cmd *exec.Cmd
	switch format {
	case "zip":
		cmd = exec.Command("zip", "-r", dest, filepath.Base(src))
		cmd.Dir = filepath.Dir(src)
	case "tar":
		cmd = exec.Command("tar", "-cf", dest, "-C", filepath.Dir(src), filepath.Base(src))
	case "tar.gz", "tgz":
		cmd = exec.Command("tar", "-czf", dest, "-C", filepath.Dir(src), filepath.Base(src))
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return cmd.Run()
}

func DecompressPath(serverID, srcPath, destPath string) error {
	base := serverDataDir(serverID)
	src := filepath.Join(base, filepath.Clean("/"+srcPath))
	dest := filepath.Join(base, filepath.Clean("/"+destPath))

	if !strings.HasPrefix(src, base) || !strings.HasPrefix(dest, base) {
		return fmt.Errorf("invalid path")
	}

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	var cmd *exec.Cmd
	lower := strings.ToLower(src)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		cmd = exec.Command("unzip", "-o", src, "-d", dest)
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		cmd = exec.Command("tar", "-xzf", src, "-C", dest)
	case strings.HasSuffix(lower, ".tar"):
		cmd = exec.Command("tar", "-xf", src, "-C", dest)
	default:
		return fmt.Errorf("unsupported archive format")
	}

	return cmd.Run()
}

func BulkCompress(serverID string, paths []string, destPath, format string) error {
	base := serverDataDir(serverID)
	dest := filepath.Join(base, filepath.Clean("/"+destPath))
	if !strings.HasPrefix(dest, base) {
		return fmt.Errorf("invalid destination")
	}

	dest = uniquePath(dest)

	var srcPaths []string
	for _, p := range paths {
		src := filepath.Join(base, filepath.Clean("/"+p))
		if !strings.HasPrefix(src, base) || src == base {
			continue
		}
		if _, err := os.Stat(src); err == nil {
			srcPaths = append(srcPaths, src)
		}
	}
	if len(srcPaths) == 0 {
		return fmt.Errorf("no valid paths")
	}

	var cmd *exec.Cmd
	switch format {
	case "zip":
		args := []string{"-r", dest}
		for _, src := range srcPaths {
			args = append(args, filepath.Base(src))
		}
		cmd = exec.Command("zip", args...)
		cmd.Dir = filepath.Dir(srcPaths[0])
	case "tar":
		args := []string{"-cf", dest, "-C", base}
		for _, src := range srcPaths {
			rel, _ := filepath.Rel(base, src)
			args = append(args, rel)
		}
		cmd = exec.Command("tar", args...)
	case "tar.gz", "tgz":
		args := []string{"-czf", dest, "-C", base}
		for _, src := range srcPaths {
			rel, _ := filepath.Rel(base, src)
			args = append(args, rel)
		}
		cmd = exec.Command("tar", args...)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return cmd.Run()
}

func uniquePath(dest string) string {
	if _, err := os.Stat(dest); err == nil {
		ext := filepath.Ext(dest)
		nameWithoutExt := strings.TrimSuffix(dest, ext)
		if strings.HasSuffix(nameWithoutExt, ".tar") {
			nameWithoutExt = strings.TrimSuffix(nameWithoutExt, ".tar")
			ext = ".tar" + ext
		}
		for i := 1; ; i++ {
			newDest := fmt.Sprintf("%s_%d%s", nameWithoutExt, i, ext)
			if _, err := os.Stat(newDest); os.IsNotExist(err) {
				return newDest
			}
		}
	}
	return dest
}
