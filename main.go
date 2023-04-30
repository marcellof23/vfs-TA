package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/afero"
)

//	func readZipFile(zf *zip.File) ([]byte, error) {
//		f, err := zf.Open()
//		if err != nil {
//			return nil, err
//		}
//		defer f.Close()
//		return io.ReadAll(f)
//	}
//
//	func main() {
//		dst := "output"
//		endpoint := "http://localhost:8080/api/v1/backups"
//
//		resp, err := http.Get(endpoint)
//		if err != nil {
//			log.Fatal(err)
//		}
//		defer resp.Body.Close()
//
//		body, err := io.ReadAll(resp.Body)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		for _, f := range zipReader.File {
//			filePath := filepath.Join(dst, f.Name)
//			fmt.Println("unzipping file ", filePath)
//
//			if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
//				fmt.Println("invalid file path")
//				return
//			}
//			if f.FileInfo().IsDir() {
//				fmt.Println("creating directory...")
//				os.MkdirAll(filePath, os.ModePerm)
//				continue
//			}
//
//			if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
//				panic(err)
//			}
//
//			dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
//			if err != nil {
//				panic(err)
//			}
//
//			fileInArchive, err := f.Open()
//			if err != nil {
//				panic(err)
//			}
//
//			if _, err := io.Copy(dstFile, fileInArchive); err != nil {
//				panic(err)
//			}
//
//			dstFile.Close()
//			fileInArchive.Close()
//		}
//	}
func main() {
	// Create a new MemMapFs filesystem
	fs := afero.NewMemMapFs()

	err := fs.MkdirAll("/a/b/c", 0o777)
	if err != nil {
		fmt.Println(err)
	}

	err = afero.WriteFile(fs, "/a/b/c/hello.txt", []byte("Hello world"), 0o600)
	if err != nil {
		fmt.Println(err)
	}

	info, err := fs.Stat("/a")
	if err != nil {
		fmt.Println(err)
	}
	if !info.Mode().IsDir() {
		fmt.Println("/a: mode is not directory")
	}
	if !info.ModTime().After(time.Now().Add(-1 * time.Hour)) {
		fmt.Printf("/a: mod time not set, got %s", info.ModTime())
	}

	if info.Mode() != os.FileMode(os.ModeDir|0o755) {
		fmt.Printf("/a: wrong permissions, expected drwxr-xr-x, got %s", info.Mode())
	}
	info, err = fs.Stat("/a/b")
	if err != nil {

		fmt.Println(err)
	}
}
