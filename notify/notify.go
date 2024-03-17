package notify

import (
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type SendMail struct {
	Email   string `json:"Email"`
	Message string `json:"Message"`
}

func InitEmailNotifier() (p *kafka.Producer) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"client.id":         "producer",
		"acks":              "all"})

	if err != nil {
		fmt.Printf("Failed to create producer: %s\n", err)
	}
	return
}

func NotifyEmailService(p *kafka.Producer, topic, recieverEmail, message string) {

	email := SendMail{
		Email:   recieverEmail,
		Message: message,
	}

	value, err := json.Marshal(email)

	if err != nil {
		fmt.Println(err)
	}
	if err := p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: 0},
		Value:          value},
		nil,
	); err != nil {
		fmt.Println("here occured an error")
		fmt.Println(err)
	}

	fmt.Println("notified")

}
