package fsys

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/marcellof23/vfs-TA/constant"
)

func (fs *Filesystem) CheckCPPath(pathSource, pathDest string) error {
	var remainingPathDest string
	splitPaths := strings.Split(pathDest, "/")
	splitPaths = splitPaths[:len(splitPaths)-1]
	remainingPathDest = filepath.ToSlash(filepath.Join(splitPaths...))
	if len(splitPaths) > 1 {
		_, err := fs.verifyPath(remainingPathDest)
		if err != nil {
			return err
		}
	}

	_, err := fs.Stat(pathSource)
	if err != nil {
		return fmt.Errorf("%s: %s", err.Error(), constant.ErrPathNotFound.Error())
	}

	_, err = fs.Stat(pathDest)
	if err == nil {
		return fmt.Errorf("%s: %s", err.Error(), constant.ErrPathNotFound.Error())
	}

	return nil
}

func (fs *Filesystem) CheckCPRecPath(pathSource, pathDest string) error {
	_, err := fs.Stat(pathSource)
	if err != nil {
		return fmt.Errorf("%s: %s", err.Error(), constant.ErrPathNotFound.Error())
	}

	_, err = fs.searchFS(pathDest)
	if err != nil {
		return fmt.Errorf("%s: %s", err.Error(), constant.ErrPathNotFound.Error())
	}

	return nil
}
