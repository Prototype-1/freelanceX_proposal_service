package kafka

import (
	"github.com/segmentio/kafka-go"
	"context"
	"encoding/json"
	"log"
)

type ProposalEvent struct {
	ProposalID   string `json:"proposal_id"`
	ClientID     string `json:"client_id"`
	FreelancerID string `json:"freelancer_id"`
	Title        string `json:"title"`
	EventType    string `json:"event_type"`
	Status       string `json:"status"`     
}

func ProduceProposalEvent(broker, topic string, event ProposalEvent) error {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer w.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = w.WriteMessages(context.Background(),
		kafka.Message{
			Value: data,
		},
	)

	if err != nil {
		log.Printf("Kafka write error: %v", err)
		return err
	}

	log.Printf("Produced event to Kafka: %+v", event)
	return nil
}
