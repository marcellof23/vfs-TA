package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	network       = "tcp"
	brokerAddress = "localhost:9092"
	topic         = "command-log"
)

var (
	commandLog *log.Logger
)

type Message struct {
	Command       string
	AbsPathSource string
	AbsPathDest   string
	Token         string
	FileMode      uint64
	Buffer        []byte
}

type Effector func(context.Context, Message) error

func Retry(effector Effector, delay time.Duration) Effector {
	return func(ctx context.Context, msg Message) error {
		log, ok := ctx.Value("logger").(*log.Logger)
		if !ok {
			return fmt.Errorf("logger not initiated")
		}

		for r := 0; ; r++ {
			err := effector(ctx, msg)
			if err == nil || r >= 20 {
				return err
			}

			log.Printf("Function call failed, retrying in %v\n", delay)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func ProduceCommand(ctx context.Context, msg Message) error {

	log, ok := ctx.Value("logger").(*log.Logger)
	if !ok {
		return fmt.Errorf("logger not initiated")
	}

	writer := kafka.Writer{
		Addr:       kafka.TCP(brokerAddress),
		BatchBytes: 1e9,
		Topic:      topic,
	}

	var buff []byte
	buff, err := json.Marshal(msg)
	if err != nil {
		log.Println("failed to marshal:", err)
		return err
	}

	err = writer.WriteMessages(
		ctx,
		kafka.Message{Value: buff},
	)

	log.Println(msg.Command, msg.AbsPathSource, msg.AbsPathDest)

	if err != nil {
		log.Println("failed to write messages:", err)
		return err
	}

	if err := writer.Close(); err != nil {
		log.Println("failed to close writer:", err)
		return err
	}

	return nil
}
