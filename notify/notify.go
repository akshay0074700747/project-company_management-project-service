package notify

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/akshay0074700747/project-company_management-project-service/helpers"
)

type SendMail struct {
	Email   string `json:"Email"`
	Message string `json:"Message"`
}

func InitEmailNotifier() (p sarama.SyncProducer) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = 5
	config.Producer.Retry.Backoff = 50 * time.Millisecond

	var err error
	for i := 0; i < 8; i++ {
		p, err = sarama.NewSyncProducer([]string{"host.docker.internal:29092"}, config)
		if err != nil {
			if i == 7 {
				log.Fatal("Closingg: %v", err)
			}
			fmt.Println("Error creating producer : ", i, ": %v", err)
			time.Sleep(time.Second * 3)
		} else {
			break
		}
	}
	return
}

func NotifyEmailService(p sarama.SyncProducer, topic, recieverEmail, message string) {

	email := SendMail{
		Email:   recieverEmail,
		Message: message,
	}

	value, err := json.Marshal(email)

	if err != nil {
		fmt.Println(err)
	}
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: 0,
		Value:     sarama.ByteEncoder(value),
	}
	_, offset, err := p.SendMessage(msg)
	if err != nil {
		helpers.PrintErr(err, "error sending message to Kafka")
		return
	}
	fmt.Println(offset, " completed...")
	fmt.Println("notified")

}
