package publisher

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
)

type Publisher struct {
	topic *pubsub.Topic
}

func InitDefault() *Publisher {
	ctx := context.Background()
	proj := "phonic-weaver-375914"
	topic := "command-topic"
	client, err := pubsub.NewClient(ctx, proj)
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}
	t := createTopicIfNotExists(client, topic)

	pubs := &Publisher{
		topic: t,
	}

	return pubs
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

func (p *Publisher) Publish(msg string) error {
	ctx := context.Background()

	result := p.topic.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})
	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("pubsub: result.Get: %v", err)
	}
	fmt.Printf("Published a message; msg ID: %v\n", id)
	return nil
}
