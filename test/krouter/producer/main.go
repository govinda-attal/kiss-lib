package main

import (
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"

	"github.com/govinda-attal/kiss-lib/pkg/kasync"
)

func main() {

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		panic(err)
	}

	defer p.Close()

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	// Produce messages to topic (asynchronously)
	topic := "Greetings"
	if err := p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(`{"name":"govinda"}`),
		Headers: []kafka.Header{
			kafka.Header{Key: kasync.MsgHdrMsgType, Value: []byte(kasync.MsgTypeEvent)},
			kafka.Header{Key: kasync.MsgHdrMsgName, Value: []byte("Welcome")},
		},
		Key: []byte(uuid.New().String()),
	}, nil); err != nil {
		log.Println(err)
	}

	if err := p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(`{"name":"govinda"}`),
		Headers: []kafka.Header{
			kafka.Header{Key: kasync.MsgHdrMsgType, Value: []byte(kasync.MsgTypeEvent)},
			kafka.Header{Key: kasync.MsgHdrMsgName, Value: []byte("Farewell")},
		},
		Key: []byte(uuid.New().String()),
	}, nil); err != nil {
		log.Println(err)
	}

	// Wait for message deliveries before shutting down
	p.Flush(15 * 1000)
}
