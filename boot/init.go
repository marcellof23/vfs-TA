package boot

import (
	"github.com/marcellof23/vfs-TA/lib/afero"
)

type MemFilesystem struct {
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
	RemoveFile(filename string) error
	RemoveDir(path string) error
	ListDir()
	Usage(comms []string) bool
	Execute(comms []string) bool
}

func InitFilesystem() *MemFilesystem {
	mfs := &afero.MemMapFs{}
	return &MemFilesystem{
		MFS: mfs,
	}
}
