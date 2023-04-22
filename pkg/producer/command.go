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

type Message struct {
	Command string
	AbsPath string
	Token   string
	Buffer  []byte
}

type Effector func(context.Context, Message) error

func Retry(effector Effector, delay time.Duration) Effector {
	return func(ctx context.Context, msg Message) error {
		log, ok := ctx.Value("logger").(*log.Logger)
		if !ok {
			return fmt.Errorf("logger not initiated")
		}

		for {
			err := effector(ctx, msg)
			if err == nil {
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
	partition := 0

	log, ok := ctx.Value("logger").(*log.Logger)
	if !ok {
		return fmt.Errorf("logger not initiated")
	}

	conn, err := kafka.DialLeader(ctx, network, brokerAddress, topic, partition)
	if err != nil {
		log.Println("failed to dial leader:", err)
		return err
	}

	var buff []byte
	if buff, err = json.Marshal(msg); err != nil {
		log.Println("failed to marshal:", err)
		return err
	}

	_, err = conn.WriteMessages(
		kafka.Message{Value: buff},
	)
	if err != nil {
		log.Println("failed to write messages:", err)
		return err
	}

	if err := conn.Close(); err != nil {
		log.Println("failed to close writer:", err)
		return err
	}

	return nil
}
