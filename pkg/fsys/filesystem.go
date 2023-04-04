package fsys

import (
	"fmt"
	"path/filepath"

	"github.com/marcellof23/vfs-TA/boot"
	"github.com/marcellof23/vfs-TA/constant"
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

// TearDown gracefully ends the current session.
func (fs *filesystem) Stat() error {
	info, err := fs.MFS.Stat(fs.rootPath)
	if err != nil {
		fmt.Println(err)
		return err
	} else {
		fmt.Println("File: ", info.Name())
		fmt.Println("Size: ", info.Size())
		fmt.Println("Access: ", info.Mode())
		fmt.Println("Modify: ", info.ModTime())
	}

	return nil
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
	if fs.rootPath == "." || path[0] != '/' {
		prefixPath = fs.rootPath + "/"
	}

	absPath := prefixPath + path
	err := fs.MFS.RemoveAll(absPath)
	if err != nil {
		return err
	}

	walkFn := func(path string, fs *filesystem, err error) error {
		delete(fs.directories, path)
		return nil
	}

	err = walkDir(fs.directories[path], path, walkFn)
	if err != nil {
		fmt.Print(err)
	}
	delete(fs.directories, path)

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

func (fs *filesystem) Testing(path string) {

	if fs.doesFileExists(path) {
		fmt.Println("XXXXXX")
	} else {
		fmt.Println("YYYYYY")
	}

	//walkFn := func(path string, fs *filesystem, err error) error {
	//	fmt.Printf("%s \n", path)
	//	return nil
	//}
	//
	//walkDir(fs, ".", walkFn)
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

// ReloadFilesys Resets the VFS and scraps all changes made up to this point.
// (basically like a rerun of InitFilesystem())
func (fs *filesystem) ReloadFilesys() {
	fmt.Println("Refreshing...")
}

// TearDown gracefully ends the current session.
func (fs *filesystem) TearDown() {
	fmt.Println("Teardown")
}
