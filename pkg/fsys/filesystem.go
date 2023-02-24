package fsys

import (
	"fmt"
	"os"
)

// A global list of all files created and their respective names for
// ease of lookup.
var globalFileTable map[uint64]string

// The data structure for each File.
type File struct {
	name     string // The name of the File.
	rootPath string // The absolute path of the File.
	fileHash uint64 // The unique File hash assigned to this File on creation.
	fileType string // The type of the File.
	content  []byte // The File's content in bytes.
	size     uint64 // The size in bytes of the File.
}

//  The core struct that makes up the filesystem's File/directory
type FileSystem struct {
	name        string                 // The name of the current directory we're in.
	rootPath    string                 // The absolute path to this directory.
	files       []File                 // The list of files in this directory.
	directories map[string]*FileSystem // The list of directories in this directory.
	prev        *FileSystem            // a reference pointer to this directory's parent directory.
}

// Root node.
var root *FileSystem

// InitFilesystem scans the current directory and builds the VFS from it.
func InitFilesystem() *FileSystem {
	// recursively grab all files and directories from this level downwards.
	root = &FileSystem{
		name:        ".",
		rootPath:    ".",
		directories: make(map[string]*FileSystem),
	}
	fs := root
	fmt.Println("Welcome to the tiny virtual filesystem.")
	return fs
}

// pwd prints the current working directory.
func (fs *FileSystem) pwd() {
	fmt.Println(fs.rootPath)
}

// reloadFilesys Resets the VFS and scraps all changes made up to this point.
// (basically like a rerun of InitFilesystem())
func (fs *FileSystem) reloadFilesys() {
	fmt.Println("Refreshing...")
}

// tearDown gracefully ends the current session.
func (fs *FileSystem) tearDown() {
	fmt.Println("Teardown")
}

// saveState aves the state of the VFS at this time.
func (fs *FileSystem) saveState() {
	fmt.Println("Save the current state of the VFS")
}

// open will allow for opening files in virtual space.
func (fs *FileSystem) open() error {
	fmt.Println("open() called")
	return nil
}

// close closes open virtual files.
func (fs *FileSystem) close() error {
	fmt.Println("close() called")
	return nil
}

// mkDir makes a virtual directory.
func (fs *FileSystem) mkDir(dirName string) bool {

	if _, exists := fs.directories[dirName]; exists {
		fmt.Println("mkdir : directory already exists")
		return false
	}

	newDir := &FileSystem{
		name:        dirName,
		rootPath:    fs.rootPath + "/" + dirName,
		directories: make(map[string]*FileSystem),
		prev:        fs,
	}
	fs.directories[dirName] = newDir
	return false
}

// removeFile removes a File from the virtual filesystem.
func (fs *FileSystem) removeFile() error {
	fmt.Println("removeFile() called")
	return nil
}

// removeDir removes a directory from the virtual filesystem.
func (fs *FileSystem) removeDir() error {
	fmt.Println("removeDir() called")
	return nil
}

// listDir lists a directory's contents.
func (fs *FileSystem) listDir() {

	if fs.files != nil {
		fmt.Println("File:")
		for _, file := range fs.files {
			fmt.Printf("\t%s\n", file.name)
		}
	}
	if len(fs.directories) > 0 {
		fmt.Println("Directories:")
		for dirName := range fs.directories {
			fmt.Printf("\t%s\n", dirName)
		}
	}
}

// usage prints verifies that each command has the correct amount of
// command arguments and throws an error if not.
func (fs *FileSystem) usage(comms []string) bool {
	switch comms[0] {
	case "mkdir":
		if len(comms) < 2 {
			fmt.Println("Usage : mkdir [list of directories to make]")
			return false
		}
	case "open":
		if len(comms) != 2 {
			fmt.Println("Usage : open [File name]")
			return false
		}
	}
	return true
}

// Execute runs the commands passed into it.
func (fs *FileSystem) Execute(comms []string) *FileSystem {
	if fs.usage(comms) == false {
		return fs
	}
	switch comms[0] {
	case "mkdir":
		fs.mkDir(comms[1])
	case "pwd":
		fs.pwd()
	case "open":
		fs.open()
	case "close":
		fs.close()
	case "ls":
		fs.listDir()
	case "rm":
		fs.removeFile()
		fs.removeDir()
	case "exit":
		fs.tearDown()
		os.Exit(1)
	default:
		fmt.Println(comms[0], ": Command not found")
	}
	return fs
}
