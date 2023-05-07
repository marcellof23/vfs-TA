package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func main() {
	// SSH client configuration
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("codegeass7359>"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Connect to the SSH server
	conn, err := ssh.Dial("tcp", "localhost:22", config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	// Open SFTP session
	client, err := sftp.NewClient(conn)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Download directory recursively
	srcDir := "<remote directory>"
	dstDir := "<local directory>"
	err = downloadDir(client, srcDir, dstDir)
	if err != nil {
		panic(err)
	}

	fmt.Println("Directory download successful!")
}

// Download directory recursively
func downloadDir(client *sftp.Client, srcDir string, dstDir string) error {
	// Get directory contents
	entries, err := client.ReadDir(srcDir)
	if err != nil {
		return err
	}

	// Create destination directory if it doesn't exist
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		os.Mkdir(dstDir, 0755)
	}

	// Download files and recurse into subdirectories
	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			// Recurse into subdirectory
			err := downloadDir(client, srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Download file
			srcFile, err := client.Open(srcPath)
			if err != nil {
				return err
			}
			defer srcFile.Close()

			dstFile, err := os.Create(dstPath)
			if err != nil {
				return err
			}
			defer dstFile.Close()

			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				return err
			}

			// Set file permissions and timestamps to match source file
			srcStat, err := srcFile.Stat()
			if err != nil {
				return err
			}
			err = os.Chmod(dstPath, srcStat.Mode())
			if err != nil {
				return err
			}
			err = os.Chtimes(dstPath, srcStat.ModTime(), srcStat.ModTime())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
