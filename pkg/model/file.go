package model

import "github.com/marcellof23/vfs-TA/boot"

type File struct {
	Name     string // The Name of the F.
	RootPath string // The absolute path of the F.
}

type FileDir struct {
	Name        string                 // The Name of the current directory we're in.
	RootPath    string                 // The absolute path to this directory.
	Files       map[string]*File       // The list of Files in this directory.
	Directories map[string]*Filesystem // The list of Directories in this directory.
	Prev        *Filesystem            // a reference pointer to this directory's parent directory.
}

type Filesystem struct {
	*boot.MemFilesystem
	*FileDir
}
