package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	// Open the file to split
	file, err := os.Open("demo-folder2/file1.bin")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Get the file size and calculate the chunk size
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	fileSize := fileInfo.Size()
	chunkSize := 20 * int64(1024*1024) // 20MB chunks
	numChunks := (fileSize + chunkSize - 1) / chunkSize

	// Create a channel to communicate between goroutines
	chunks := make(chan []byte)

	// Spawn goroutines to read chunks of the file
	for i := int64(0); i < numChunks; i++ {
		go func(offset int64) {
			// Seek to the offset of the chunk
			_, err := file.Seek(offset, io.SeekStart)
			if err != nil {
				panic(err)
			}

			// Read the chunk
			chunk := make([]byte, chunkSize)
			n, err := file.Read(chunk)
			if err != nil && err != io.EOF {
				panic(err)
			}

			// Send the chunk to the channel
			chunks <- chunk[:n]
		}(i * chunkSize)
	}

	// Create a new file to write the chunks to
	outFile, err := os.Create("split_file.bin")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	// Write each chunk to the new file
	for i := int64(0); i < numChunks; i++ {
		chunk := <-chunks
		_, err := outFile.Write(chunk)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Chunk %d written\n", i+1)
	}
}
