package fsys

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// our Shell object.
type Shell struct {
	Fs *Filesystem
}

// InitShell initializes our Shell object.
func InitShell(fs *Filesystem) *Shell {
	return &Shell{
		Fs: fs,
	}
}

func (s *Shell) SetFilesystem(fs *Filesystem) {
	s.Fs = fs
}

// ClearScreen clears the terminal screen.
func (s *Shell) ClearScreen() {
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
func (s *Shell) ChDir(dirName string) {
	if dirName == "/" {
		s.Fs = root
		return
	}

	fsVerified, err := s.verifyPath(dirName)
	if err != nil {
		return
	}
	s.Fs = fsVerified
}

func (s *Shell) cat(filename string) {
	segments := strings.Split(filename, "/")
	if len(segments) == 1 {
		if _, exists := s.Fs.files[filename]; exists {
			s.readFile(s.Fs.files[filename].rootPath)
		} else {
			fmt.Println("cat : file doesn't exist")
		}
	} else {
		dirPath := s.reassemble(segments)
		tmp, _ := s.verifyPath(dirPath)

		if _, exists := tmp.files[segments[len(segments)-1]]; exists {
			s.readFile(tmp.files[segments[len(segments)-1]].rootPath)
			fmt.Println("File exists")
		} else {
			fmt.Println("cat : file doesn't exist")
		}
	}
}

func (s *Shell) usage(comms []string) bool {
	switch comms[0] {
	case "cd":
		if len(comms) != 2 {
			fmt.Println("Usage : cd [target directory")
			return true
		}
	case "cat":
		if len(comms) != 2 {
			fmt.Println("Usage : Cat [target file]")
			return true
		}
	}
	return true
}

func (s *Shell) Execute(comms []string) bool {
	if s.usage(comms) == false {
		return false
	}
	switch comms[0] {
	case "cd":
		s.ChDir(comms[1])
	//case "cat":
	//	s.cat(comms[1])
	case "clear":
		s.ClearScreen()
	default:
		return false
	}
	return true
}
