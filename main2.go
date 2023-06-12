package main

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"google.golang.org/api/option"

	"github.com/marcellof23/vfs-TA/pkg/pubsub_notify"
)

func main() {
	client, err := pubsub.NewClient(context.Background(), "phonic-weaver-375914", option.WithCredentialsFile("credentials.json"))
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}

	i := 0

	t := pubsub_notify.GetTopic(context.Background(), client, "command-topic")

	for i = 0; i < 20; i++ {
		_, err := client.CreateSubscription(context.Background(), "command-"+uuid.New().String(), pubsub.SubscriptionConfig{
			ExpirationPolicy: 72 * time.Hour,
			Topic:            t,
			AckDeadline:      20 * time.Second,
		})
		if err != nil {
			log.Fatalf("Could not create pubsub Client: %v", err)
		}
	}

}
