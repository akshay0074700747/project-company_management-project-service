package notify

import (
	"fmt"

	"github.com/akshay0074700747/proto-files-for-microservices/pb"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

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

	email := pb.Email{
		Reciever: recieverEmail,
		Message:  message,
	}
	value, err := Serialize(&email)
	if err != nil {
		fmt.Println(err)
	}
	if err := p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          value},
		nil,
	); err != nil {
		fmt.Println("here occured an error")
		fmt.Println(err)
	}

	fmt.Println("notified")

}

func Serialize(m protoreflect.ProtoMessage) ([]byte, error) {

	serialized, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	return serialized, nil
}
