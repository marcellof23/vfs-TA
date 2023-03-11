package fsys

import (
	"fmt"
	"os"

	"github.com/marcellof23/vfs-TA/boot"
)

type filesystem struct {
	*boot.Filesystem
}

func New(fs *boot.Filesystem) boot.FilesystemIntf {
	return &filesystem{fs}
}

// pwd prints pwd() the current working directory.
func (fs *filesystem) Pwd() {
	fmt.Println("")
}

// ReloadFilesys Resets the VFS and scraps all changes made up to this point.
// (basically like a rerun of InitFilesystem())
func (fs *filesystem) ReloadFilesys() {
	fmt.Println("Refreshing...")
}

// TearDown gracefully ends the current session.
func (fs *filesystem) TearDown() {
	fmt.Println("Teardown")
}

func (fs *filesystem) Cat(file string) {
	fmt.Println("")
}

func (fs *filesystem) Touch(filename string) bool {
	return true
}

// SaveState aves the state of the VFS at this time.
func (fs *filesystem) SaveState() {
	fmt.Println("Save the current state of the VFS")
}

// Open will allow for opening files in virtual space.
func (fs *filesystem) Open() error {
	fmt.Println("Open() called")
	return nil
}

// Close closes Open virtual files.
func (fs *filesystem) Close() error {
	fmt.Println("Close() called")
	return nil
}

// MkDir makes a virtual directory.
func (fs *filesystem) MkDir(dirName string) bool {
	err := fs.MFS.Mkdir(file, 0o700)
	if err != nil {
		fmt.Println(err)
	}
	info, err := fs.MFS.Stat(file)
	fmt.Println(info.Mode().String())
	return true
}

// RemoveFile removes a File from the virtual filesystem.
func (fs *filesystem) RemoveFile() error {
	fmt.Println("RemoveFile() called")
	return nil
}

const file = "hello"

// RemoveDir removes a directory from the virtual filesystem.
func (fs *filesystem) RemoveDir(path string) error {
	err := fs.MFS.RemoveAll(path)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// ListDir lists a directory's contents.
func (fs *filesystem) ListDir() {
	fs.MFS.List()
}

// Usage prints verifies that each command has the correct amount of
// command arguments and throws an error if not.
func (fs *filesystem) Usage(comms []string) bool {
	switch comms[0] {
	case "mkdir":
		if len(comms) < 2 {
			fmt.Println("Usage : mkdir [list of directories to make]")
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
func (fs *filesystem) Execute(comms []string) bool {
	if fs.Usage(comms) == false {
		return false
	}
	switch comms[0] {
	case "mkdir":
		fs.MkDir(comms[1])
	case "pwd":
		fs.Pwd()
	case "Open":
		fs.Open()
	case "Close":
		fs.Close()
	case "ls":
		fs.ListDir()
	case "rm":
		fs.RemoveFile()
		fs.RemoveDir(comms[1])
	case "exit":
		fs.TearDown()
		os.Exit(1)
	default:
		fmt.Println(comms[0], ": Command not found")
		return false
	}
	return true
}
