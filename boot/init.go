package boot

import "github.com/spf13/afero"

var (
	FS  afero.Fs
	AFS *afero.Afero
)

type Filesystem struct {
	AFS *afero.Afero
}

func InitFilesystem() Filesystem {
	FS = afero.NewMemMapFs()
	AFS = &afero.Afero{Fs: FS}
	return Filesystem{
		AFS: AFS,
	}
}
