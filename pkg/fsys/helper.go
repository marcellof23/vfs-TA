package fsys

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/edsrzf/mmap-go"
	"github.com/spf13/afero"

	"github.com/marcellof23/vfs-TA/boot"
)

type WalkDirFunc func(path string, d fileDir, err error) error

func New() *filesystem {
	// uncomment for recursively grab all files and directories from this level downwards.
	root = replicateFilesystem(".", nil)

	// root = makeFilesystem(".", ".", nil)
	fsys := root
	return fsys
}

// testFilessytemCreation initializes the filesystem by replicating
// the current root directory and all it's child direcctories.
func replicateFilesystem(dirName string, fs *filesystem) *filesystem {
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
		fi, _ = os.Stat(dirName + "//" + fileName.Name())
		mode := fi.Mode()
		if mode.IsDir() {
			fs.directories[fileName.Name()] = makeFilesystem(fileName.Name(), strings.ReplaceAll(dirName, "//", "/")+"/"+fileName.Name(), fs)
			err := fs.MFS.MkdirAll(fileName.Name(), 0o700)
			if err != nil {
				fmt.Println(err)
			}
			replicateFilesystem(dirName+"//"+fileName.Name(), fs.directories[fileName.Name()])
		} else {
			fs.files[fileName.Name()] = &file{
				name:     fileName.Name(),
				rootPath: strings.ReplaceAll(dirName, "//", "/") + "/" + fileName.Name(),
			}

			fs.MFS.Create(fs.files[fileName.Name()].rootPath)
		}

		index++
	}
	return fs
}

func makeFilesystem(dirName string, rootPath string, prev *filesystem) *filesystem {
	fs := boot.InitFilesystem()
	return &filesystem{
		fs,
		&fileDir{
			name:        dirName,
			rootPath:    rootPath,
			files:       make(map[string]*file),
			directories: make(map[string]*filesystem),
			prev:        prev,
		},
	}
}

func walkDir(filesys *boot.Filesystem, fs afero.Fs, root string, walkFn filepath.WalkFunc) error {
	// Open the root directory.
	f, err := fs.Open(root)
	if err != nil {
		return err
	}
	defer f.Close()

	// Read the file info for the root directory.
	info, err := f.Stat()
	if err != nil {
		return err
	}

	// If the root directory is a file, call the walk function with its info.
	if !info.IsDir() {
		return walkFn(root, info, nil)
	}

	// Otherwise, walk through the root directory and call the walk function for each file or subdirectory.
	return afero.Walk(fs, root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil // skip directories
		}

		// Open the file using mmap.
		f, err := fs.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		osFile, err := os.Open(f.Name())
		if err != nil {
			fmt.Println(err)
		}
		defer osFile.Close()

		// Map the file to memory.
		data, err := mmap.Map(osFile, mmap.RDONLY, 0)
		if err != nil {
			return err
		}
		defer data.Unmap()
		// Call the walk function with the file info and mmap data.
		return walkFn(path, info, err)
	})
}
