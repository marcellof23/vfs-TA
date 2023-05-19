package fsys

import (
	"container/list"
	"fmt"
	"os"
)

var (
	LruCache        *LRUCache
	MemoryThreshold = int64(30 * 1024)
)

type Node struct {
	Filename string
	FileSize int64
	KeyPtr   *list.Element
}

type LRUCache struct {
	Queue     *list.List
	Items     map[string]*Node
	TotalSize int64
}

func Constructor() *LRUCache {
	return &LRUCache{Queue: list.New(), Items: make(map[string]*Node), TotalSize: 0}
}

func (l *LRUCache) Put(key string, value int64, content []byte, fs *Filesystem) {
	removedStat, _ := fs.MFS.Stat(key)
	removedFile, _ := fs.MFS.OpenFile(key, os.O_RDWR|os.O_TRUNC, removedStat.Mode())
	defer removedFile.Close()

	if item, ok := l.Items[key]; !ok {
		FileSizeMap[key] = value
		if l.TotalSize >= MemoryThreshold {
			back := l.Queue.Back()
			l.Queue.Remove(back)
			delete(l.Items, back.Value.(string))

			filename := back.Value.(string)
			destStat, _ := fs.MFS.Stat(filename)
			destFile, _ := fs.MFS.OpenFile(filename, os.O_RDWR|os.O_TRUNC, destStat.Mode())
			defer destFile.Close()

			l.TotalSize -= FileSizeMap[filename]
			destFile.Truncate(0)
			destFile.Write([]byte{})
		}
		l.TotalSize += value
		l.Items[key] = &Node{FileSize: value, Filename: key, KeyPtr: l.Queue.PushFront(key)}
		if removedStat.Size() == 0 {
			removedFile.Truncate(value)
			removedFile.Write(content)
		}
	} else {
		item.Filename = key
		item.FileSize = value
		l.TotalSize += value
		l.Items[key] = item
		l.Queue.MoveToFront(item.KeyPtr)
		if removedStat.Size() == 0 {
			removedFile.Truncate(value)
			removedFile.Write(content)
		}
	}
}

func (l *LRUCache) Get(key string) int64 {
	if item, ok := l.Items[key]; ok {
		l.Queue.MoveToFront(item.KeyPtr)
		return item.FileSize
	}
	return -1
}

func (l *LRUCache) PrintCache() int64 {
	for item, maps := range l.Items {
		fmt.Printf("%v %v \n", item, maps)
	}

	fmt.Println("Total Size: ", l.TotalSize)
	return -1
}

func main() {
	fmt.Println("halo")
}
