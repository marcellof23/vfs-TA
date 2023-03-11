package fsys

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"

	"github.com/marcellof23/vfs-TA/boot"
)

// our shell object.
type shell struct {
	Fs  *boot.Filesystem
	env map[string]string // the environment varialbes.
}

// Env variable stores current directory infomration.
var env map[string]string

// InitShell initializes our shell object.
func InitShell(fs *boot.Filesystem) *shell {
	env = make(map[string]string)
	return &shell{
		Fs:  fs,
		env: env,
	}
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
	fmt.Println("")
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
