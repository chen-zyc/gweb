package gweb

import (
	"net/http"
	"os"
)

type EmptyDirFile struct {
	http.File
}

// 覆盖默认实现，不返回子文件
func (edf EmptyDirFile) Readdir(_ int) ([]os.FileInfo, error) { return nil, nil }

type OnlyFilesFS struct {
	fs http.FileSystem
}

func (offs OnlyFilesFS) Open(name string) (http.File, error) {
	f, err := offs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return EmptyDirFile{f}, nil
}
