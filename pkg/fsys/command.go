package fsys

import (
	"context"
	"fmt"
	"os"

	"github.com/marcellof23/vfs-TA/constant"
)

// Filesystem Commands

// Usage prints verifies that each command has the correct amount of
// command arguments and throws an error if not.
func (fs *Filesystem) Usage(comms []string) bool {
	switch comms[0] {
	case "mkdir":
		if len(comms) < 2 {
			fmt.Println(constant.UsageCommandMkdir)
			return false
		}
	case "pwd":
		if len(comms) > 1 {
			fmt.Println(constant.UsageCommandPwd)
			return false
		}
	case "ls":
		if len(comms) > 1 {
			fmt.Println(constant.UsageCommandLs)
			return false
		}
	case "cat":
		if len(comms) < 2 {
			fmt.Println(constant.UsageCommandCat)
			return false
		}
	case "stat":
		if len(comms) < 2 {
			fmt.Println(constant.UsageCommandStat)
			return false
		}
	case "rm":
		if len(comms) < 2 {
			fmt.Println(constant.UsageCommandRm)
			return false
		}
	case "cp":
		if len(comms) < 3 {
			fmt.Println(constant.UsageCommandCp)
			return false
		}
	case "chmod":
		if len(comms) < 3 {
			fmt.Println(constant.UsageCommandChmod)
			return false
		}
	case "upload":
		if len(comms) < 3 {
			fmt.Println(constant.UsageCommandUpload)
			return false
		}
	case "migrate":
		if len(comms) < 3 {
			fmt.Println(constant.UsageCommandMigrate)
			return false
		}
	case "download":
		if len(comms) < 3 {
			fmt.Println(constant.UsageCommandDownload)
			return false
		}
	}
	return true
}

// Execute runs the commands passed into it.
func (fs *Filesystem) Execute(ctx context.Context, comms []string, publish bool) bool {
	var err error
	if fs.Usage(comms) == false {
		return false
	}

	role, ok := ctx.Value("role").(string)
	if !ok {
		fmt.Println("User is not authorized!")
		return false
	}

	switch comms[0] {
	case "mkdir":
		err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.MkDir, ctx, publish, comms[1])
	case "pwd":
		fs.Pwd()
	case "ls":
		fs.ListDir()
	case "cat":
		err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.Cat, ctx, publish, comms[1])
	case "stat":
		stat, errs := fs.Stat(comms[1])
		if err == nil {
			fs.PrintStat(stat, comms[1])
		}
		err = errs
	case "rm":
		if comms[1] == "-r" {
			err = fs.FilesystemAccessAuth(ctx, role, true, comms[0], fs.RemoveDir, ctx, publish, comms[2])
		} else {
			err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.RemoveFile, ctx, publish, comms[1])
		}
	case "cp":
		if comms[1] == "-r" {
			err = fs.FilesystemAccessAuth(ctx, role, true, comms[0], fs.CopyDir, ctx, publish, comms[2], comms[3])
		} else {
			err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.CopyFile, ctx, publish, comms[1], comms[2])
		}
	case "chmod":
		err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.Chmod, ctx, publish, comms[1], comms[2])
	case "upload":
		if comms[1] == "-r" {
			err = fs.FilesystemAccessAuth(ctx, role, true, comms[0], fs.UploadDir, ctx, publish, comms[2], comms[3])
		} else {
			err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.UploadFile, ctx, publish, comms[1], comms[2])
		}
	case "migrate":
		err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.Migrate, ctx, publish, comms[1], comms[2])
	case "download":

		if comms[1] == "-r" {
			err = fs.FilesystemAccessAuth(ctx, role, true, comms[0], fs.DownloadRecursive, ctx, publish, comms[2], comms[3])
		} else {
			err = fs.FilesystemAccessAuth(ctx, role, false, comms[0], fs.DownloadFile, ctx, publish, comms[1], comms[2])
		}
	case "test":
		fs.Testing(ctx, comms[1])
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

// Shell Commands

func (s *Shell) Usage(comms []string) bool {
	switch comms[0] {
	case "cd":
		if len(comms) != 2 {
			fmt.Println("Usage : cd [target directory]")
			return false
		}
	}
	return true
}

func (s *Shell) Execute(ctx context.Context, comms []string) bool {
	role, ok := ctx.Value("role").(string)
	if !ok {
		fmt.Println("User is not authorized!")
		return false
	}

	if s.Usage(comms) == false {
		return true
	}
	switch comms[0] {
	case "cd":
		err := s.Fs.FilesystemAccessAuth(ctx, role, false, comms[0], s.ChDir, ctx, false, comms[1])
		if err != nil {
			fmt.Println(err.Error())
		}
	case "clear":
		s.ClearScreen()
	default:
		return false
	}
	return true
}
