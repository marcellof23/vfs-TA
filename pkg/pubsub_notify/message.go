package pubsub_notify

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub"
)

type MessageCommand struct {
	ClientID    string
	FullCommand string
	FileMode    uint64
	Uid         int
	Gid         int
	Buffer      []byte
}

func GetTopic(ctx context.Context, c *pubsub.Client, topic string) *pubsub.Topic {
	log, ok := ctx.Value("server-logger").(*log.Logger)
	if !ok {
		return &pubsub.Topic{}
	}

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
