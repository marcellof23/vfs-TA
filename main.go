package main

import (
	"container/list"
	"fmt"
	"os"

	"github.com/marcellof23/vfs-TA/lib/afero"
)

var (
	LruCache        LRUCache
	MemoryThreshold = int64(100)
)

type Node struct {
	FileSize int64
	Filename string

	KeyPtr *list.Element
}

type LRUCache struct {
	Queue     *list.List
	Items     map[string]*Node
	TotalSize int64
}

func Constructor() *LRUCache {
	return &LRUCache{Queue: list.New(), Items: make(map[string]*Node), TotalSize: 0}
}

func (l *LRUCache) Put(key string, value int64) {
	if item, ok := l.Items[key]; !ok {
		if l.TotalSize >= MemoryThreshold {
			back := l.Queue.Back()
			l.Queue.Remove(back)

			delete(l.Items, back.Value.(string))

			//filename := back.Value.(int64)
			//destStat, _ := fs.MFS.Stat(filename)
			//destFile, _ := fs.MFS.OpenFile(filename, os.O_RDWR|os.O_CREATE, destStat.Mode())
			//destFile.Truncate(0)
			//destFile.Write([]byte{})
		}
		l.TotalSize += value
		l.Items[key] = &Node{FileSize: value, Filename: key, KeyPtr: l.Queue.PushFront(key)}
	} else {
		item.Filename = key
		item.FileSize = value
		l.Items[key] = item
		l.Queue.MoveToFront(item.KeyPtr)
	}
}

func (l *LRUCache) Get(key string) int64 {
	if item, ok := l.Items[key]; ok {
		l.Queue.MoveToFront(item.KeyPtr)
		return item.FileSize
	}
	return -1
}

func main() {
	//obj := Constructor()
	//obj.Put("filename1", 50)
	//obj.Put("filename2", 55)
	//obj.Put("filename3", 55)
	//fmt.Println(obj.Get("filename1"))

	fs := &afero.MemMapFs{}
	fs.Create("test")
	afero.WriteFile(fs, "test", []byte("hellosss"), 0o775)

	info, _ := fs.Stat("test")
	fmt.Println(info.Size())

	f, _ := fs.OpenFile("test", os.O_RDWR|os.O_TRUNC, 0o775)
	b := make([]byte, 10000)
	f.Read(b)
	fmt.Println(string(b))

	f.Write([]byte{})
	f.Close()

	info, _ = fs.Stat("test")
	fmt.Println(info.Size())
}
