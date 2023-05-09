package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/marcellof23/vfs-TA/boot"
	"github.com/marcellof23/vfs-TA/constant"
)

const (
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
	Offset        int
	Uid           int
	Gid           int
	Buffer        []byte
}

type Effector func(context.Context, Message) error

func Retry(effector Effector, delay time.Duration) Effector {
	return func(ctx context.Context, msg Message) error {
		log, ok := ctx.Value("server-logger").(*log.Logger)
		if !ok {
			return fmt.Errorf("ERROR: logger not initiated")
		}

		for r := 0; ; r++ {
			err := effector(ctx, msg)
			if err == nil || r >= 20 {
				return err
			}

			log.Printf("ERROR: Function call failed, retrying in %v\n", delay)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func IntermediateHealthCheck(ctx context.Context, dep *boot.Dependencies) error {
	healthURL := constant.Protocol + dep.Config().Server.Addr + "/health"

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		resp, err := http.Get(healthURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			fmt.Println("\nServer is down, your changed may not be saved!")
			ticker2 := time.NewTicker(3 * time.Second)
			for range ticker2.C {
				resp, err := http.Get(healthURL)
				if err == nil && resp.StatusCode == http.StatusOK {
					fmt.Println("\nServer is ready, your can continue!")
					ticker2.Stop()
					break
				}
			}
		}
	}

	return nil
}

func KafkaHealthCheck(ctx context.Context) error {
	dialer := &kafka.Dialer{
		Timeout:  10 * time.Second,
		ClientID: "health-check-client",
	}
	readerConfig := kafka.ReaderConfig{
		Brokers:  []string{brokerAddress},
		Topic:    topic,
		Dialer:   dialer,
		MinBytes: 1,
		MaxBytes: 10e6,
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if checkHealth(readerConfig) != nil {
			fmt.Println("\nServer is down, your changed may not be saved!")
			ticker2 := time.NewTicker(3 * time.Second)
			for range ticker2.C {
				if checkHealth(readerConfig) == nil {
					fmt.Println("\nServer is ready, your can continue!")
					ticker2.Stop()
					break
				}
			}
		}
	}

	return nil
}

func checkHealth(conf kafka.ReaderConfig) error {
	r := kafka.NewReader(conf)
	_, err := r.FetchMessage(context.Background())
	defer r.Close()
	if err != nil {
		return err
	} else {
		return nil
	}
}

func ProduceCommand(ctx context.Context, msg Message) error {

	log, ok := ctx.Value("server-logger").(*log.Logger)
	if !ok {
		return fmt.Errorf("ERROR: logger not initiated")
	}

	writer := kafka.Writer{
		Addr:       kafka.TCP(brokerAddress),
		BatchBytes: 1e9,
		Topic:      topic,
	}

	var buff []byte
	buff, err := json.Marshal(msg)
	if err != nil {
		log.Println("ERROR: failed to marshal:", err)
		return err
	}

	err = writer.WriteMessages(
		ctx,
		kafka.Message{Value: buff},
	)

	log.Println(msg.Command, msg.AbsPathSource, msg.AbsPathDest, msg.Uid, msg.Gid, msg.FileMode)

	if err != nil {
		log.Println("ERROR: kafka writer failed to write messages:", err)
		return err
	}

	if err := writer.Close(); err != nil {
		log.Println("ERROR: failed to close writer:", err)
		return err
	}

	return nil
}
