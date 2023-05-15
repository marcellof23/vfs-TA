package main

import (
	"fmt"
	"os"
)

func main() {
	// Open the file to split
	d1 := []byte("hello\ngo\n")
	d2 := []byte("hello\ngo2\n")
	err := os.WriteFile("output-folder/test", d1, 0644)
	if err != nil {
		fmt.Println(err)
	}

	file, _ := os.OpenFile("output-folder/test", os.O_RDWR, 0644)
	d1len := len(d1)

	_, err = file.WriteAt(d2, int64(d1len)-2)
	if err != nil {
		fmt.Println(err)
	}

}
