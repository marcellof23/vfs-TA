package producer

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/segmentio/kafka-go"
)

const (
	network       = "tcp"
	brokerAddress = "localhost:9092"
	topic         = "command-log"
)

type Message struct {
	Command string
	Buffer  []byte
}

func ProduceCommand(ctx context.Context, msg Message) {
	f, err := os.OpenFile("testlogfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	l := log.New(f, "kafka writer: ", 0)

	partition := 0
	conn, err := kafka.DialLeader(ctx, network, brokerAddress, topic, partition)
	if err != nil {
		l.Println("failed to dial leader:", err)
	}

	var buff []byte
	if buff, err = json.Marshal(msg); err != nil {
		l.Println("failed to marshal:", err)
	}

	_, err = conn.WriteMessages(
		kafka.Message{Value: buff},
	)

	if err != nil {
		log.Fatal("failed to write messages:", err)
	}

	if err := conn.Close(); err != nil {
		l.Fatal("failed to close writer:", err)
	}
}
