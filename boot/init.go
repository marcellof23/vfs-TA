package boot

import (
	"github.com/spf13/afero"
)

type Filesystem struct {
	MFS *afero.MemMapFs
}

type FilesystemIntf interface {
	Pwd()
	ReloadFilesys()
	TearDown()
	Cat(file string)
	Touch(filename string) bool
	SaveState()
	Open() error
	Close() error
	MkDir(dirName string) bool
	RemoveFile() error
	RemoveDir(path string) error
	ListDir()
	Usage(comms []string) bool
	Execute(comms []string) bool
}

func InitFilesystem() *Filesystem {
	mfs := &afero.MemMapFs{}
	return &Filesystem{
		MFS: mfs,
	}
}
