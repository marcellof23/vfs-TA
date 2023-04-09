package fsys

import (
	"fmt"
	"strings"

	"github.com/spf13/afero"
)

func (s *shell) doesDirExist(dirName string, fs *Filesystem) bool {
	if _, found := fs.directories[dirName]; found {
		return true
	}
	return false
}

func (s *shell) verifyPath(dirName string) (*Filesystem, error) {
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
			fmt.Printf("Error : %s doesn't exist\n", dirName)
			return s.Fs, fmt.Errorf("Error : %s doesn't exist\n", dirName)
		}
	}
	return checker, nil
}

func (s *shell) handleRootNav(dirName string) *Filesystem {
	if dirName[0] == '/' {
		return root
	}
	return s.Fs
}

func (s *shell) reassemble(dirPath []string) string {
	counter := 1
	var finishedPath string

	finishedPath = dirPath[0]
	for counter < len(dirPath)-1 {
		finishedPath = finishedPath + "/" + dirPath[counter]
		counter++
	}
	return finishedPath
}

func (s *shell) readFile(filename string) {
	data, _ := afero.ReadFile(s.Fs.MFS, filename)
	fmt.Println(filename)
	fmt.Println(string(data))
}
