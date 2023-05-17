package subscriber

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
)

type Subscriber struct {
	Subs *pubsub.Subscription
}

func InitDefault() *Subscriber {
	ctx := context.Background()
	proj := "phonic-weaver-375914"
	topic := "command-topic"
	subsID := "command-sub"
	client, err := pubsub.NewClient(ctx, proj)
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}
	_ = createTopicIfNotExists(client, topic)

	sub := client.Subscription(subsID)

	subs := &Subscriber{
		Subs: sub,
	}

	return subs
}

func createTopicIfNotExists(c *pubsub.Client, topic string) *pubsub.Topic {
	ctx := context.Background()
	t := c.Topic(topic)
	ok, err := t.Exists(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if ok {
		return t
	}
	t, err = c.CreateTopic(ctx, topic)
	if err != nil {
		log.Fatalf("Failed to create the topic: %v", err)
	}
	return t
}

func (s *Subscriber) ListenMessage(ctx context.Context) error {
	err := s.Subs.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		// Process received message
		fmt.Printf("Received message: %s\n", string(msg.Data))

		// Acknowledge the message to remove it from the subscription
		msg.Ack()
	})
	if err != nil {
		fmt.Printf("Error receiving messages: %v\n", err)
	}

	return nil
}
