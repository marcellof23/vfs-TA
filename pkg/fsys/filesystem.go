package fsys

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"

	"github.com/marcellof23/vfs-TA/boot"
)

type file struct {
	name     string // The name of the file.
	rootPath string // The absolute path of the file.
}

type fileDir struct {
	name        string                 // The name of the current directory we're in.
	rootPath    string                 // The absolute path to this directory.
	files       map[string]*file       // The list of files in this directory.
	directories map[string]*filesystem // The list of directories in this directory.
	prev        *filesystem            // a reference pointer to this directory's parent directory.
}

type filesystem struct {
	*boot.Filesystem
	*fileDir
}

// Root node.
var root *filesystem

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
	if _, exists := fs.files[file]; exists {
		fileObj, _ := fs.MFS.Open(fs.rootPath + "/" + file)
		var fileObjContent []byte
		cnt, err := fileObj.Read(fileObjContent)
		if cnt != 0 && err != nil {
			fmt.Println(string(fileObjContent))
		}
	} else {
		fmt.Println("cat : file doesn't exist")
	}
}

func (fs *filesystem) Touch(filename string) bool {
	fs.MFS.Create(fs.rootPath + "/" + filename)
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
	newDir := &fileDir{
		name:        dirName,
		rootPath:    filepath.Join(fs.rootPath, dirName),
		files:       make(map[string]*file),
		directories: make(map[string]*filesystem),
		prev:        fs,
	}

	fs.directories[dirName] = &filesystem{fs.Filesystem, newDir}
	return true
}

// RemoveFile removes a File from the virtual filesystem.
func (fs *filesystem) RemoveFile(filename string) error {
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
	delete(fs.files, filename)

	return nil
}

// RemoveDir removes a directory from the virtual filesystem.
func (fs *filesystem) RemoveDir(path string) error {
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
	delete(fs.directories, path)

	return nil
}

// ListDir lists a directory's contents.
func (fs *filesystem) ListDir() {
	//if fs.files != nil {
	//	for _, file := range fs.files {
	//		fmt.Printf("%s\n", file.name)
	//	}
	//}
	//if len(fs.directories) > 0 {
	//	for dirName := range fs.directories {
	//		fmt.Println(constant.ColorBlue, dirName)
	//	}
	//	fmt.Print(constant.ColorReset)
	//}

	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}
		fmt.Println(path)
		return nil
	}

	// Use afero.Walk to get a list of all files and directories in the file system.
	// Then, use filepath.Walk to perform the actual traversal and call walkFn for each file.
	if err := afero.Walk(fs.MFS, ".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}
		return filepath.Walk(path, walkFn)
	}); err != nil {
		fmt.Println(err)
		return
	}

	//walkFn := func(path string, info os.FileInfo, err error) error {
	//	fmt.Printf("%s (%d bytes)\n", path, info.Size())
	//	return nil
	//}
	//if err := walkDir(fs.Filesystem, fs.MFS, ".", walkFn); err != nil {
	//	fmt.Println(err)
	//}
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
	case "cat":
		if len(comms) < 2 {
			fmt.Println("Usage : cat [list of directories to make]")
			return false
		}
	case "rm":
		if len(comms) < 2 {
			fmt.Println("Usage : rm [File name]")
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
	case "cat":
		fs.Cat(comms[1])
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
