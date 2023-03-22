package fsys

import (
	"fmt"
	"os"
	"runtime"

	"github.com/marcellof23/vfs-TA/boot"
	"github.com/marcellof23/vfs-TA/constant"
)

type file struct {
	name     string // The name of the file.
	rootPath string // The absolute path of the file.
}

type fileSystem struct {
	name        string                 // The name of the current directory we're in.
	rootPath    string                 // The absolute path to this directory.
	files       map[string]*file       // The list of files in this directory.
	directories map[string]*filesystem // The list of directories in this directory.
	prev        *filesystem            // a reference pointer to this directory's parent directory.
}

type filesystem struct {
	*boot.Filesystem
	*fileSystem
}

// Root node.
var root *filesystem

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func New(fs *boot.Filesystem) *filesystem {

	root = makeFilesystem(".", ".", nil)
	root.Filesystem = fs
	fsys := root
	return fsys
}

func makeFilesystem(dirName string, rootPath string, prev *filesystem) *filesystem {
	return &filesystem{
		nil,
		&fileSystem{
			name:        dirName,
			rootPath:    rootPath,
			directories: make(map[string]*filesystem),
			prev:        prev,
		},
	}
}

// pwd prints pwd() the current working directory.
func (fs *filesystem) Pwd() {
	fmt.Println(fs.rootPath)
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
	err := fs.MFS.MkdirAll(dirName, 0o700)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if _, exists := fs.directories[dirName]; exists {
		fmt.Println("mkdir : directory already exists")
		return false
	}
	newDir := &fileSystem{
		name:        dirName,
		rootPath:    fs.rootPath + "/" + dirName,
		files:       make(map[string]*file),
		directories: make(map[string]*filesystem),
		prev:        fs,
	}

	fs.directories[dirName] = &filesystem{fs.Filesystem, newDir}
	return true
}

// RemoveFile removes a File from the virtual filesystem.
func (fs *filesystem) RemoveFile(filename string) error {
	delete(fs.files, filename)

	var prefixPath string
	if fs.rootPath == "." {
		prefixPath = fs.rootPath + "/"
	}

	absPath := prefixPath + filename
	err := fs.MFS.Remove(absPath)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// RemoveDir removes a directory from the virtual filesystem.
func (fs *filesystem) RemoveDir(path string) error {
	delete(fs.directories, path)

	var prefixPath string
	if fs.rootPath == "." {
		prefixPath += "/"
	}

	absPath := prefixPath + path
	err := fs.MFS.RemoveAll(absPath)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// ListDir lists a directory's contents.
func (fs *filesystem) ListDir() {
	if fs.files != nil {
		for _, file := range fs.files {
			fmt.Printf("%s\n", file.name)
		}
	}
	if len(fs.directories) > 0 {
		for dirName := range fs.directories {
			fmt.Println(constant.ColorBlue, dirName)
		}
		fmt.Print(constant.ColorReset)
	}
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
		fs.RemoveFile(comms[1])
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
