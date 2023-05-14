package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"sync"
)

var (
	cnt       = 0
	partsDir  = "output-folder"
	linesSize = 4 * 1024
	chunkSz   = 2 * 1024
)

func chunkBytes(data []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}
	return chunks
}
func ProcessChunk(chunk []byte, linesPool, stringPool *sync.Pool, partition int) {

	var wg sync.WaitGroup

	datas := stringPool.Get().(string)
	//fmt.Println(len(chunk), "AAA")
	datas = string(chunk)

	linesPool.Put(chunk)

	datasSlice := chunkBytes([]byte(datas), chunkSz)

	stringPool.Put(datas)

	chunkSize := chunkSz
	n := len(datasSlice)
	noOfThread := n / chunkSize

	if n%chunkSize != 0 {
		noOfThread++
	}

	fmt.Println(noOfThread)

	for i := 0; i < (noOfThread); i++ {
		wg.Add(1)

		fmt.Println(i*chunkSize, int(math.Min(float64((i+1)*chunkSize), float64(len(datasSlice)))), "Ei")
		go func(s int, e int) {
			defer wg.Done() //to avoid deadlocks
			for j := s; j < e; j++ {
				text := datasSlice[j]

				filename := filepath.Join(partsDir, fmt.Sprintf("file-%d", partition+j))
				os.Create(filename)
				os.WriteFile(filename, text, 0666)
				fmt.Println(filename, s, e)

				if len(text) == 0 {
					continue
				}
			}
		}(i*chunkSize, int(math.Min(float64((i+1)*chunkSize), float64(len(datasSlice)))))
	}

	//fmt.Println(cnt, "AHA")

	wg.Wait()
	datasSlice = nil
}

func Process(f *os.File) error {

	linesPool := sync.Pool{New: func() interface{} {
		lines := make([]byte, linesSize)
		return lines
	}}

	stringPool := sync.Pool{New: func() interface{} {
		lines := ""
		return lines
	}}

	r := bufio.NewReader(f)

	var wg sync.WaitGroup
	var mutex sync.Mutex

	partition := 0
	for {
		buf := linesPool.Get().([]byte)

		n, err := r.Read(buf)
		buf = buf[:n]

		if n == 0 {
			if err != nil {
				fmt.Println(err)
				break
			}
			if err == io.EOF {
				break
			}
			return err
		}

		wg.Add(1)

		go func(partition int) {
			ProcessChunk(buf, &linesPool, &stringPool, partition)
			wg.Done()
		}(partition)

		mutex.Lock()
		partition += 2
		mutex.Unlock()

	}

	wg.Wait()
	return nil
}

func main() {
	filename := "pkg/fsys/filesystem.go"
	//filename := "demo-folder2/file5.bin"

	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("cannot able to read the file", err)
		return
	}

	if _, err := os.Stat(partsDir); os.IsNotExist(err) {
		if err = os.MkdirAll(partsDir, os.ModePerm); err != nil {
			log.Fatal("error creating directory:", err)
		}
	}

	Process(file)
}
