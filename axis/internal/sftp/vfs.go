package sftp

import (
	"cauthon-axis/internal/server"
	"io"
	"os"

	"github.com/pkg/sftp"
	"github.com/spf13/afero"
)

type vfsHandler struct {
	serverID string
}

type listerat []os.FileInfo

func (f listerat) ListAt(ls []os.FileInfo, offset int64) (int, error) {
	if offset >= int64(len(f)) {
		return 0, io.EOF
	}
	n := copy(ls, f[offset:])
	if n < len(ls) {
		return n, io.EOF
	}
	return n, nil
}

func (h *vfsHandler) Filelist(r *sftp.Request) (sftp.ListerAt, error) {
	fs := server.GetVFS(h.serverID)
	switch r.Method {
	case "List":
		entries, err := afero.ReadDir(fs, r.Filepath)
		if err != nil {
			return nil, err
		}
		var ret []os.FileInfo
		for _, e := range entries {
			ret = append(ret, e)
		}
		return listerat(ret), nil
	case "Stat":
		info, err := fs.Stat(r.Filepath)
		if err != nil {
			return nil, err
		}
		return listerat([]os.FileInfo{info}), nil
	case "Readlink":
		return nil, sftp.ErrSSHFxOpUnsupported
	}
	return nil, sftp.ErrSSHFxOpUnsupported
}

func (h *vfsHandler) Filecmd(r *sftp.Request) error {
	fs := server.GetVFS(h.serverID)
	switch r.Method {
	case "Setstat":
		return nil
	case "Rename":
		return fs.Rename(r.Filepath, r.Target)
	case "Rmdir":
		return fs.Remove(r.Filepath)
	case "Remove":
		return fs.Remove(r.Filepath)
	case "Mkdir":
		return fs.MkdirAll(r.Filepath, 0755)
	case "Symlink":
		return sftp.ErrSSHFxOpUnsupported
	}
	return sftp.ErrSSHFxOpUnsupported
}

func (h *vfsHandler) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	fs := server.GetVFS(h.serverID)
	f, err := fs.Open(r.Filepath)
	if err != nil {
		return nil, err
	}
	return f.(io.ReaderAt), nil
}

func (h *vfsHandler) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	fs := server.GetVFS(h.serverID)
	flags := 0
	pflags := r.Pflags()
	
	if pflags.Append {
		flags |= os.O_APPEND
	}
	if pflags.Creat {
		flags |= os.O_CREATE
	}
	if pflags.Trunc {
		flags |= os.O_TRUNC
	}
	if pflags.Read && pflags.Write {
		flags |= os.O_RDWR
	} else if pflags.Read {
		flags |= os.O_RDONLY
	} else if pflags.Write {
		flags |= os.O_WRONLY
	}

	f, err := fs.OpenFile(r.Filepath, flags, 0644)
	if err != nil {
		return nil, err
	}
	return f.(io.WriterAt), nil
}

func newVFSHandlers(serverID string) sftp.Handlers {
	h := &vfsHandler{serverID: serverID}
	return sftp.Handlers{
		FileGet:  h,
		FilePut:  h,
		FileCmd:  h,
		FileList: h,
	}
}
