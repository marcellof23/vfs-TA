package fsys

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/marcellof23/vfs-TA/constant"
)

var (
	ErrUnauthorizedAccess = errors.New("you are not authorized to perform this action")
	ErrPathNotFound       = errors.New("no such file or directory")
)

func concludeAccess(accessSlice []string) string {
	r := true
	w := true
	x := true
	for _, v := range accessSlice {
		strAccess := strings.Split(v, "")
		r = r && (strAccess[0] == "r")
		w = w && (strAccess[1] == "w")
		x = x && (strAccess[2] == "x")
	}

	var access string
	if r {
		access += "r"
	} else {
		access += "-"
	}

	if w {
		access += "w"
	} else {
		access += "-"
	}

	if x {
		access += "x"
	} else {
		access += "-"
	}
	return access
}

func (fs *Filesystem) getAccess(path, role string) (string, error) {
	checker := fs.handleRootNav(path)
	segments := strings.Split(path, "/")

	var accessSlice []string
	rootStat, _ := fs.MFS.Stat(".")
	rootAccess := rootStat.Mode().Perm().String()

	if role == "Normal" {
		rootAccess = rootAccess[len(rootAccess)-3:]
	} else {
		rootAccess = rootAccess[1:4]
	}
	accessSlice = append(accessSlice, rootAccess)

	for _, segment := range segments {
		if segment == "." {
			continue
		}
		if len(segment) == 0 {
			continue
		}
		if segment == ".." {
			if checker.prev == nil {
				continue
			}
			checker = checker.prev
		} else if fs.doesDirExistRelativePath(segment, checker) {
			dirName := filepath.ToSlash(filepath.Join(checker.rootPath, segment))
			dirStat, _ := fs.MFS.Stat(dirName)
			dirAccess := dirStat.Mode().Perm().String()

			if role == "Normal" {
				dirAccess = dirAccess[len(dirAccess)-3:]
			} else {
				dirAccess = dirAccess[1:4]
			}
			accessSlice = append(accessSlice, dirAccess)
			checker = checker.directories[segment]
		} else if fs.doesFileExistRelativePath(segment, checker) {
			filename := filepath.ToSlash(filepath.Join(checker.rootPath, segment))
			fileStat, _ := fs.MFS.Stat(filename)
			fileAccess := fileStat.Mode().Perm().String()
			if role == "Normal" {
				fileAccess = fileAccess[len(fileAccess)-3:]
			} else {
				fileAccess = fileAccess[1:4]
			}
			accessSlice = append(accessSlice, fileAccess[len(fileAccess)-3:])
			acc := concludeAccess(accessSlice)
			if acc == "" {
				return "", ErrUnauthorizedAccess
			}
			return acc, nil
		} else {
			acc := concludeAccess(accessSlice)
			if acc == "" {
				return "", ErrUnauthorizedAccess
			}
			return acc, nil
		}
	}
	acc := concludeAccess(accessSlice)
	if acc == "" {
		return "", ErrUnauthorizedAccess
	}
	return acc, nil
}

func checkAccess(access, requiredAccess string) bool {
	reqAccess := strings.Split(requiredAccess, "")
	acc := strings.Split(access, "")
	for i := 0; i < 3; i++ {
		if reqAccess[i] != "-" && acc[i] != reqAccess[i] {
			return false
		}
	}
	return true
}

func CallFunc(v reflect.Value, vargs []reflect.Value) error {
	res := v.Call(vargs)

	if res[0].Interface() == nil {
		return nil
	}

	err := res[0].Interface().(error)
	if err != nil {
		return err
	}
	return nil
}

func (fs *Filesystem) FilesystemAccessAuth(role string, isRec bool, command string, f interface{}, args ...interface{}) error {
	v := reflect.ValueOf(f)
	vargs := make([]reflect.Value, len(args))
	for i, arg := range args {
		vargs[i] = reflect.ValueOf(arg)
	}

	var comms = make([]string, 3)
	sourceOnly := false
	comms[1] = args[1].(string)
	if len(args) > 2 {
		comms[2] = args[2].(string)
	} else {
		sourceOnly = true
	}

	if _, ok := constant.Command[command]; ok {
		var srcPath, dstPath string
		srcPath = comms[1]
		if !sourceOnly {
			dstPath = comms[2]
		}

		var srcAccess, dstAccess string
		srcAccess = "---"
		dstAccess = "---"
		srcAccess, err := fs.getAccess(srcPath, role)
		if err != nil {
			return err
		}

		if !sourceOnly {
			dstAccess, err = fs.getAccess(dstPath, role)
			if err != nil && dstPath != "" {
				return err
			}
		}

		fmt.Println(srcAccess, dstAccess)
		fmt.Println(srcPath, dstPath)
		switch command {
		case "cp":
			if isRec {
				err = fs.CheckCPRecPath(srcPath, dstPath)
				if err != nil {
					return err
				}

				if !checkAccess(srcAccess, "r-x") {
					return ErrUnauthorizedAccess
				}

				if !checkAccess(dstAccess, "r-x") {
					return ErrUnauthorizedAccess
				}
			} else {
				err = fs.CheckCPPath(srcPath, dstPath)
				if err != nil {
					return err
				}

				if !checkAccess(srcAccess, "r--") {
					return ErrUnauthorizedAccess
				}

				if !checkAccess(dstAccess, "-w-") {
					return ErrUnauthorizedAccess
				}
			}
		case "rm":
			if isRec {
				if !checkAccess(srcAccess, "-wx") {
					return ErrUnauthorizedAccess
				}
			} else {
				if !checkAccess(srcAccess, "-wx") {
					return ErrUnauthorizedAccess
				}
			}
		case "upload":
			if isRec {
				if !checkAccess(srcAccess, "-wx") {
					return ErrUnauthorizedAccess
				}
			} else {
				if !checkAccess(srcAccess, "-wx") {
					return ErrUnauthorizedAccess
				}
			}
		case "mkdir":
			if isRec {
				if !checkAccess(srcAccess, "-wx") {
					return ErrUnauthorizedAccess
				}
			} else {
				if !checkAccess(srcAccess, "-wx") {
					return ErrUnauthorizedAccess
				}
			}
		case "cat":
			if isRec {
				if !checkAccess(srcAccess, "-wx") {
					return ErrUnauthorizedAccess
				}
			} else {
				if !checkAccess(srcAccess, "-wx") {
					return ErrUnauthorizedAccess
				}
			}
		case "cd":
			if !checkAccess(srcAccess, "--x") {
				return ErrUnauthorizedAccess
			}
		case "chmod":
			if role == "Normal" {
				return ErrUnauthorizedAccess
			}
		}
	}

	return CallFunc(v, vargs)
}