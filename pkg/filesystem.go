package pkg

import (
	"github.com/spf13/afero"

	"github.com/marcellof23/vfs-TA/boot"
	_ "github.com/marcellof23/vfs-TA/boot"
)

func Create(name string) (afero.File, error) {
	return boot.FS.Create(name)
}
