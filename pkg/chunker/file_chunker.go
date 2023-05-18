package chunker

import (
	"bufio"
	"context"
	"io"
	"math"
	"os"
	"sync"

	"github.com/marcellof23/vfs-TA/pkg/producer"
)

var (
	pipesSize = 100 * 1024 * 1024
	chunkSz   = 25 * 1024 * 1024
)

type FileChunk struct {
	Ctx           context.Context
	Command       string
	AbsPathSource string
	AbsPathDest   string
	Token         string
	Uid           int
	Gid           int
}

func (fc *FileChunk) chunkBytes(data []byte, chunkSize int) [][]byte {
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
func (fc *FileChunk) ProcessChunk(chunk []byte, linesPool, stringPool *sync.Pool, partition int) {

	var wg sync.WaitGroup

	datas := stringPool.Get().(string)
	datas = string(chunk)

	linesPool.Put(chunk)

	datasSlice := fc.chunkBytes([]byte(datas), chunkSz)

	stringPool.Put(datas)

	chunkSize := chunkSz
	n := len(datasSlice)
	noOfThread := n / chunkSize

	if n%chunkSize != 0 {
		noOfThread++
	}

	for i := 0; i < (noOfThread); i++ {
		wg.Add(1)

		go func(s int, e int) {
			defer wg.Done() //to avoid deadlocks
			for j := s; j < e; j++ {
				data := datasSlice[j]

				msg := producer.Message{
					Command:       "write",
					Token:         fc.Token,
					AbsPathSource: fc.AbsPathSource,
					AbsPathDest:   fc.AbsPathDest,
					Buffer:        data,
					Order:         partition + j,
					Uid:           fc.Uid,
					Gid:           fc.Gid,
				}

				producer.ProduceCommand(fc.Ctx, msg)
				//fmt.Printf("%+v\n", msg)
				//r := producer.Retry(producer.ProduceCommand, 3e9)
				//go r(fc.Ctx, msg)

				if len(data) == 0 {
					continue
				}
			}
		}(i*chunkSize, int(math.Min(float64((i+1)*chunkSize), float64(len(datasSlice)))))
	}

	wg.Wait()
	datasSlice = nil
}

func (fc *FileChunk) Process(f *os.File) error {

	linesPool := sync.Pool{New: func() interface{} {
		lines := make([]byte, pipesSize)
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
				break
			}
			if err == io.EOF {
				break
			}
			return err
		}

		wg.Add(1)

		go func(partition int) {
			fc.ProcessChunk(buf, &linesPool, &stringPool, partition)
			wg.Done()
		}(partition)

		mutex.Lock()
		partition += pipesSize / chunkSz
		mutex.Unlock()

	}

	wg.Wait()
	return nil
}
