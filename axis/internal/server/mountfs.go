package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"
)

type mountFs struct {
	base   afero.Fs
	mounts map[string]afero.Fs
}

func (m *mountFs) resolve(name string) (afero.Fs, string) {
	name = filepath.Clean("/" + name)
	var bestPrefix string
	var bestFs afero.Fs

	for prefix, fs := range m.mounts {
		if name == prefix || strings.HasPrefix(name, prefix+"/") {
			if len(prefix) > len(bestPrefix) {
				bestPrefix = prefix
				bestFs = fs
			}
		}
	}

	if bestFs != nil {
		rel, _ := filepath.Rel(bestPrefix, name)
		return bestFs, filepath.Clean("/" + rel)
	}

	return m.base, name
}

func (m *mountFs) Create(name string) (afero.File, error) {
	fs, resolved := m.resolve(name)
	return fs.Create(resolved)
}

func (m *mountFs) Mkdir(name string, perm os.FileMode) error {
	fs, resolved := m.resolve(name)
	return fs.Mkdir(resolved, perm)
}

func (m *mountFs) MkdirAll(path string, perm os.FileMode) error {
	fs, resolved := m.resolve(path)
	return fs.MkdirAll(resolved, perm)
}

func (m *mountFs) Open(name string) (afero.File, error) {
	fs, resolved := m.resolve(name)
	return fs.Open(resolved)
}

func (m *mountFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	fs, resolved := m.resolve(name)
	return fs.OpenFile(resolved, flag, perm)
}

func (m *mountFs) Remove(name string) error {
	fs, resolved := m.resolve(name)
	return fs.Remove(resolved)
}

func (m *mountFs) RemoveAll(path string) error {
	fs, resolved := m.resolve(path)
	return fs.RemoveAll(resolved)
}

func (m *mountFs) Rename(oldname, newname string) error {
	fs1, res1 := m.resolve(oldname)
	fs2, res2 := m.resolve(newname)
	if fs1 != fs2 {
		return fmt.Errorf("cross-mount rename not supported")
	}
	return fs1.Rename(res1, res2)
}

func (m *mountFs) Stat(name string) (os.FileInfo, error) {
	fs, resolved := m.resolve(name)
	return fs.Stat(resolved)
}

func (m *mountFs) Name() string { return "mountFs" }

func (m *mountFs) Chmod(name string, mode os.FileMode) error {
	fs, resolved := m.resolve(name)
	return fs.Chmod(resolved, mode)
}

func (m *mountFs) Chown(name string, uid, gid int) error {
	fs, resolved := m.resolve(name)
	return fs.Chown(resolved, uid, gid)
}

func (m *mountFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	fs, resolved := m.resolve(name)
	return fs.Chtimes(resolved, atime, mtime)
}
