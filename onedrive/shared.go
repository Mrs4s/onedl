package onedrive

import (
	"onedl/models"
	"os"
)

type IShared interface {
	Init() bool
	GetRootFiles() []SharedFile
	GetFilesByPath(path string) []SharedFile
}

type SharedFile struct {
	Owner     IShared
	Name      string
	Path      string
	Size      int64
	IsDir     bool
	SpItemUrl string

	GetDownloadLink func(file SharedFile) string
}

func (file SharedFile) GetLocalTree(localPath string) map[string]SharedFile {
	if !models.PathExists(localPath) {
		_ = os.Mkdir(localPath, os.ModePerm)
	}
	result := make(map[string]SharedFile)
	if !file.IsDir {
		result[models.CombinePaths(localPath, file.Name)] = file
		return result
	}
	for _, childrenFile := range file.GetChildren() {
		for path, node := range childrenFile.GetLocalTree(models.CombinePaths(localPath, file.Name)) {
			result[path] = node
		}
	}

	return result
}

func (file SharedFile) GetChildren() []SharedFile {
	return file.Owner.GetFilesByPath(file.Path)
}
