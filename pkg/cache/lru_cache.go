package cache

import (
	"container/list"
	"fmt"
	"os"

	"github.com/marcellof23/vfs-TA/pkg/model"
)

var (
	LruCache        *LRUCache
	MemoryThreshold = int64(100)
)

type Node struct {
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

func (l *LRUCache) Put(key string, value int64, fs *model.Filesystem) {
	if item, ok := l.Items[key]; !ok {
		if l.TotalSize >= MemoryThreshold {
			back := l.Queue.Back()
			l.Queue.Remove(back)
			delete(l.Items, back.Value.(string))

			filename := back.Value.(string)
			destStat, _ := fs.MFS.Stat(filename)
			destFile, _ := fs.MFS.OpenFile(filename, os.O_RDWR|os.O_CREATE, destStat.Mode())
			destFile.Truncate(0)
			destFile.Write([]byte{})
		}
		l.TotalSize += value
		l.Items[key] = &Node{FileSize: value, KeyPtr: l.Queue.PushFront(key)}
	} else {
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
	fmt.Println("halo")
}
