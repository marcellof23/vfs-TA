package fsys

import (
	"context"
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
func (s *Shell) ChDir(_ context.Context, dirName string) error {
	if dirName == "/" {
		s.Fs = root
		return nil
	}

	fsVerified, err := s.verifyPath(dirName)
	if err != nil {
		return err
	}
	s.Fs = fsVerified
	return nil
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
