package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

// A global list of all files created and their respective names for
// ease of lookup.
var globalFileTable map[uint64]string

// The data structure for each file.
type file struct {
	name     string // The name of the file.
	rootPath string // The absolute path of the file.
	fileHash uint64 // The unique file hash assigned to this file on creation.
	fileType string // The type of the file.
	content  []byte // The file's content in bytes.
	size     uint64 // The size in bytes of the file.
}

//  The core struct that makes up the filesystem's file/directory
type fileSystem struct {
	name        string                 // The name of the current directory we're in.
	rootPath    string                 // The absolute path to this directory.
	files       map[string]*file       // The list of files in this directory.
	directories map[string]*fileSystem // The list of directories in this directory.
	prev        *fileSystem            // a reference pointer to this directory's parent directory.
}

// Root node.
var root *fileSystem

func makeFilesystem(dirName string, rootPath string, prev *fileSystem) *fileSystem {
	return &fileSystem{
		name:        dirName,
		rootPath:    rootPath,
		files:       make(map[string]*file),
		directories: make(map[string]*fileSystem),
		prev:        prev,
	}
}

func testFilesystemCreation(dirName string, fs *fileSystem) *fileSystem {
	var fi os.FileInfo
	var fileName os.FileInfo

	if dirName == "." {
		root = makeFilesystem(".", ".", nil)
		fs = root
	}
	index := 0
	files, _ := ioutil.ReadDir(dirName)
	for index < len(files) {
		fileName = files[index]
		fi, _ = os.Stat(dirName + "\\" + fileName.Name())
		mode := fi.Mode()
		if mode.IsDir() {
			fs.directories[fileName.Name()] = makeFilesystem(fileName.Name(), dirName+"/"+fileName.Name(), fs)
			testFilesystemCreation(dirName+"\\"+fileName.Name(), fs.directories[fileName.Name()])
		} else {
			fs.files[fileName.Name()] = &file{
				name:     fileName.Name(),
				rootPath: dirName + "/" + fileName.Name(),
			}
		}
		index++
	}
	return fs
}

// initFilesystem scans the current directory and builds the VFS from it.
func initFilesystem() *fileSystem {
	// recursively grab all files and directories from this level downwards.
	root = testFilesystemCreation(".", nil)
	fs := root
	fmt.Println("Welcome to the tiny virtual filesystem.")
	return fs
}

// pwd prints the current working directory.
func (fs *fileSystem) pwd() {
	fmt.Println(fs.rootPath)
}

// reloadFilesys Resets the VFS and scraps all changes made up to this point.
// (basically like a rerun of initFilesystem())
func (fs *fileSystem) reloadFilesys() {
	fmt.Println("Refreshing...")
}

// tearDown gracefully ends the current session.
func (fs *fileSystem) tearDown() {
	fmt.Println("Teardown")
}

func (fs *fileSystem) cat(file string) {
	if _, exists := fs.files[file]; exists {
		fileObj := fs.files[file]
		if len(fileObj.content) == 0 {
			data, err := ioutil.ReadFile(fs.rootPath)
			if err != nil {
				fmt.Println("cat : unable to open file")
				return
			}
			fmt.Println(string(data))
		}
	} else {
		fmt.Println("cat : file doesn't exist")
	}
}

func (fs *fileSystem) touch(filename string) bool {
	if _, exists := fs.files[filename]; exists {
		fmt.Printf("touch : file already exists")
		return false
	}
	newFile := &file{
		name:     filename,
		rootPath: fs.rootPath + "/" + filename,
	}
	fs.files[filename] = newFile
	return true
}

// saveState aves the state of the VFS at this time.
func (fs *fileSystem) saveState() {
	fmt.Println("Save the current state of the VFS")
}

// open will allow for opening files in virtual space.
func (fs *fileSystem) open() error {
	fmt.Println("open() called")
	return nil
}

// close closes open virtual files.
func (fs *fileSystem) close() error {
	fmt.Println("close() called")
	return nil
}

// mkDir makes a virtual directory.
func (fs *fileSystem) mkDir(dirName string) bool {
	if _, exists := fs.directories[dirName]; exists {
		fmt.Println("mkdir : directory already exists")
		return false
	}
	newDir := &fileSystem{
		name:        dirName,
		rootPath:    fs.rootPath + "/" + dirName,
		files:       make(map[string]*file),
		directories: make(map[string]*fileSystem),
		prev:        fs,
	}
	fs.directories[dirName] = newDir
	return true
}

// removeFile removes a file from the virtual filesystem.
func (fs *fileSystem) removeFile() error {
	fmt.Println("removeFile() called")
	return nil
}

// removeDir removes a directory from the virtual filesystem.
func (fs *fileSystem) removeDir() error {
	fmt.Println("removeDir() called")
	return nil
}

// listDir lists a directory's contents.
func (fs *fileSystem) listDir() {

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
func (fs *fileSystem) usage(comms []string) bool {
	switch comms[0] {
	case "mkdir":
		if len(comms) < 2 {
			fmt.Println("Usage : mkdir [list of directories to make]")
			return false
		}
	case "open":
		if len(comms) != 2 {
			fmt.Println("Usage : open [file name]")
			return false
		}
	}
	return true
}

// execute runs the commands passed into it.
func (fs *fileSystem) execute(comms []string) (*fileSystem, bool) {
	if fs.usage(comms) == false {
		return fs, false
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
		return fs, false
	}
	return fs, true
}
