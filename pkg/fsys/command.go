package fsys

import (
	"fmt"
	"os"

	"github.com/marcellof23/vfs-TA/constant"
)

// Usage prints verifies that each command has the correct amount of
// command arguments and throws an error if not.
func (fs *Filesystem) Usage(comms []string) bool {
	switch comms[0] {
	case "mkdir":
		if len(comms) < 2 {
			fmt.Println("Usage : mkdir [list of directories to make]")
			return false
		}
	case "cat":
		if len(comms) < 2 {
			fmt.Println("Usage : cat [list of directories to make]")
			return false
		}
	case "rm":
		if len(comms) < 2 {
			fmt.Println(constant.UsageCommandRM)
			return false
		}
	case "cp":
		if len(comms) < 4 {
			fmt.Println(constant.UsageCommandRM)
			return false
		}
	case "Open":
		if len(comms) != 2 {
			fmt.Println("Usage : Open [File name]")
			return false
		}
	}

	return true
}

// Execute runs the commands passed into it.
func (fs *Filesystem) Execute(comms []string) bool {
	var err error
	if fs.Usage(comms) == false {
		return false
	}
	switch comms[0] {
	case "mkdir":
		fs.MkDir(comms[1])
	case "pwd":
		fs.Pwd()
	case "ls":
		fs.ListDir()
	case "test":
		fs.Testing(comms[1])
	case "stat":
		stat, errs := fs.Stat(comms[1])
		if err == nil {
			fs.PrintStat(stat, comms[1])
		}
		err = errs
	case "touch":
		err = fs.Touch(comms[1])
	case "rm":
		if comms[1] == "-r" {
			err = fs.RemoveDir(comms[2])
		} else {
			err = fs.RemoveFile(comms[1])
		}
	case "cp":
		if comms[1] == "-r" {
			fs.CopyDir(comms[2], comms[3])
		} else {
			fs.CopyDir(comms[1], comms[2])
		}
	case "chmod":
		fs.Chmod(comms[1], comms[2])
	case "exit":
		fs.TearDown()
		os.Exit(1)
	default:
		fmt.Println(comms[0], ": Command not found")
		return false
	}
	if err != nil {
		fmt.Println(err.Error())
	}
	return true
}
