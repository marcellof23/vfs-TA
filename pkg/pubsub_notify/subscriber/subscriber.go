package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"google.golang.org/api/option"

	"github.com/marcellof23/vfs-TA/boot"
	"github.com/marcellof23/vfs-TA/constant"
	"github.com/marcellof23/vfs-TA/global"
	"github.com/marcellof23/vfs-TA/pkg/fsys"
	"github.com/marcellof23/vfs-TA/pkg/pubsub_notify"
)

type Subscriber struct {
	Subs *pubsub.Subscription
}

func InitDefault(ctx context.Context, dep *boot.Dependencies) (*Subscriber, error) {
	log, ok := ctx.Value("server-logger").(*log.Logger)
	if !ok {
		return &Subscriber{}, fmt.Errorf("failed to get logger from context")
	}

	proj := dep.Config().Pubsub.Project
	topic := dep.Config().Pubsub.Topic
	credentialsFile := dep.Config().Pubsub.CredentialFile

	client, err := pubsub.NewClient(ctx, proj, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}

	t := pubsub_notify.GetTopic(ctx, client, topic)

	sub, err := client.CreateSubscription(ctx, "command-"+uuid.New().String(), pubsub.SubscriptionConfig{
		ExpirationPolicy: 24 * time.Hour,
		Topic:            t,
		AckDeadline:      20 * time.Second,
	})
	if err != nil {
		log.Fatalf("Could not create pubsub Client: %v", err)
	}

	subs := &Subscriber{
		Subs: sub,
	}

	return subs, nil
}

func (s *Subscriber) ListenMessage(ctx context.Context) error {
	log, ok := ctx.Value("server-logger").(*log.Logger)
	if !ok {
		return fmt.Errorf("failed to get logger from context")
	}

	err := s.Subs.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var msgCmd pubsub_notify.MessageCommand
		if err := json.Unmarshal(msg.Data, &msgCmd); err != nil {
			log.Println("failed to unmarshal:", err)
		}

		clientID, _ := fsys.GetClientIDFromContext(ctx)
		// Message is just for other clients
		if msgCmd.ClientID != clientID {
			comms := strings.Split(msgCmd.FullCommand, " ")
			fmt.Println(comms)
			if _, ok := constant.CommandPubsub[comms[0]]; ok {
				if comms[0] == "upload-sync" {
					global.Filesys.UploadSyncFile(ctx, msgCmd)
				} else {
					global.Filesys.Execute(ctx, comms, false)
				}
			}
		}

		msg.Ack()
	})
	if err != nil {
		fmt.Printf("Error receiving messages: %v\n", err)
	}

	return nil
}
