// OB-Agent provides basic file storage can caching.
// For example, you can download a rpm file in one HTTP request, and install it in another.
// OB-Agent stores all files under /tmp directory.
// Files are temporarily stored in this directory and will be cleaned by host periodically.
// OB-Agent do not expose absolute path of files to user.
// Instead, user can only take relative paths under /tmp directory.
// For example, when user refer to file path rpms/a.rpm，the real file path is /tmp/rpms/a.rpm.
// This prevents destruction of host files by user.

package file

import (
	"path/filepath"
	"strings"
)

const defaultBasePath = "/tmp"

type Path struct {
	BasePath string // file base path
	RelPath  string // file relative path
}

func (p *Path) FullPath() string {
	if strings.HasPrefix(p.RelPath, defaultBasePath) {
		return p.RelPath
	}
	return filepath.Join(p.BasePath, p.RelPath)
}

func (p *Path) String() string {
	return p.FullPath()
}

func (p *Path) FileName() string {
	return filepath.Base(p.RelPath)
}

func (p *Path) Join(elem string) *Path {
	return NewPathFromRelPath(filepath.Join(p.RelPath, elem))
}

func NewPathFromRelPath(relPath string) *Path {
	return &Path{
		BasePath: defaultBasePath,
		RelPath:  relPath,
	}
}
