package fsys

import (
	"fmt"
	"strings"

	"github.com/marcellof23/vfs-TA/lib/afero"

	"github.com/marcellof23/vfs-TA/constant"
)

func (s *Shell) doesDirExist(dirName string, fs *Filesystem) bool {
	if _, found := fs.directories[dirName]; found {
		return true
	}
	return false
}

func (s *Shell) verifyPath(dirName string) (*Filesystem, error) {
	checker := s.handleRootNav(dirName)
	segments := strings.Split(dirName, "/")

	for _, segment := range segments {
		if len(segment) == 0 {
			continue
		}
		if segment == ".." {
			if checker.prev == nil {
				continue
			}
			checker = checker.prev
		} else if s.doesDirExist(segment, checker) == true {
			checker = checker.directories[segment]
		} else {
			return s.Fs, constant.Errorf(constant.ErrPathFormatNotFound.Error(), dirName)
			//return s.Fs, fmt.Errorf("Error : %s doesn't exist", dirName)
		}
	}
	return checker, nil
}

func (s *Shell) handleRootNav(dirName string) *Filesystem {
	if dirName[0] == '/' {
		return root
	}
	return s.Fs
}

func (s *Shell) reassemble(dirPath []string) string {
	counter := 1
	var finishedPath string

	finishedPath = dirPath[0]
	for counter < len(dirPath)-1 {
		finishedPath = finishedPath + "/" + dirPath[counter]
		counter++
	}
	return finishedPath
}

func (s *Shell) readFile(filename string) {
	data, _ := afero.ReadFile(s.Fs.MFS, filename)
	fmt.Println(string(data))
}
