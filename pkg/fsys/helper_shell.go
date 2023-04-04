package fsys

import (
	"fmt"
	"io/ioutil"
	"strings"
)

func (s *shell) doesDirExist(dirName string, fs *filesystem) bool {
	if _, found := fs.directories[dirName]; found {
		return true
	}
	return false
}

func (s *shell) verifyPath(dirName string) (*filesystem, error) {
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

func (s *shell) handleRootNav(dirName string) *filesystem {
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
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	fmt.Println(string(dat))
}
