package fsys

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// our shell object.
type shell struct {
	Fs *filesystem
}

// InitShell initializes our shell object.
func InitShell(fs *filesystem) *shell {
	return &shell{
		Fs: fs,
	}
}

func (s *shell) SetFilesystem(fs *filesystem) {
	s.Fs = fs
}

// ClearScreen clears the terminal screen.
func (s *shell) ClearScreen() {
	clear := make(map[string]func())
	clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	cls, ok := clear[runtime.GOOS]
	if ok {
		cls()
	}
}

// ChDir switches you to a different active directory.
func (s *shell) ChDir(dirName string) {
	if dirName == "/" {
		s.Fs = root
		return
	}
	s.Fs = s.verifyPath(dirName)
}

func (s *shell) doesDirExist(dirName string, fs *filesystem) bool {
	if _, found := fs.directories[dirName]; found {
		return true
	}
	return false
}

func (s *shell) verifyPath(dirName string) *filesystem {
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
			return s.Fs
		}
	}
	return checker
}

func (s *shell) handleRootNav(dirName string) *filesystem {
	if dirName[0] == '/' {
		return root
	}
	return s.Fs
}

func (s *shell) reassemble(dirPath []string) string {
	return ""
}

func (s *shell) readFile(filename string) {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	fmt.Println(string(dat))
}

func (s *shell) cat(filename string) {
	fmt.Println("")
}

func (s *shell) usage(comms []string) bool {
	switch comms[0] {
	case "cd":
		if len(comms) != 2 {
			fmt.Println("Usage : cd [target directory")
			return true
		}
	case "Cat":
		if len(comms) != 2 {
			fmt.Println("Usage : Cat [target file]")
			return true
		}
	}
	return true
}

func (s *shell) Execute(comms []string) bool {

	if s.usage(comms) == false {
		return false
	}
	switch comms[0] {
	case "cd":
		s.ChDir(comms[1])
	case "Cat":
		s.cat(comms[1])
	case "clear":
		s.ClearScreen()
	default:
		return false
	}
	return true
}
