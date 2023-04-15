package producer

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	network       = "tcp"
	brokerAddress = "localhost:9092"
	topic         = "command-log"
)

func ProduceCommand(ctx context.Context) {
	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	l := log.New(os.Stdout, "kafka writer: ", 0)

	partition := 0
	i := 0

	conn, err := kafka.DialLeader(ctx, network, brokerAddress, topic, partition)
	if err != nil {
		l.Fatal("failed to dial leader:", err)
	}

	for {
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		_, err = conn.WriteMessages(
			kafka.Message{Value: []byte("this is message " + strconv.Itoa(i))},
		)

		if err != nil {
			log.Fatal("failed to write messages:", err)
		}

		time.Sleep(time.Second)
	}

	if err := conn.Close(); err != nil {
		l.Fatal("failed to close writer:", err)
	}
}
