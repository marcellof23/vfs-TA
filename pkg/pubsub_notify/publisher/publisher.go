package publisher

import (
	"context"
	"encoding/json"
	"log"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"

	"github.com/marcellof23/vfs-TA/boot"
	"github.com/marcellof23/vfs-TA/pkg/pubsub_notify"
)

type Publisher struct {
	topic *pubsub.Topic
}

func InitDefault(ctx context.Context, dep *boot.Dependencies) (*Publisher, error) {
	proj := dep.Config().Pubsub.Project
	topic := dep.Config().Pubsub.Topic
	credentialsFile := dep.Config().Pubsub.CredentialFile

	client, err := pubsub.NewClient(ctx, proj, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}
	t := pubsub_notify.GetTopic(ctx, client, topic)

	pubs := &Publisher{
		topic: t,
	}

	return pubs, nil
}

func (p *Publisher) Publish(ctx context.Context, msg pubsub_notify.MessageCommand) error {
	var buff []byte
	buff, err := json.Marshal(msg)
	if err != nil {
		log.Println("ERROR: failed to marshal:", err)
		return err
	}

	_ = p.topic.Publish(ctx, &pubsub.Message{
		Data: buff,
	})

	return nil
}
