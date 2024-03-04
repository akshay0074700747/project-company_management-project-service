package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/akshay0074700747/project-company_management-project-service/entities"
	"github.com/akshay0074700747/project-company_management-project-service/helpers"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func (project *ProjectServiceServer) StartConsuming() {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  "localhost:9092",
		"group.id":           "taskConsumers",
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
		"allow.auto.create.topics": true})
	if err != nil {
		helpers.PrintErr(err, "error occured at creating a kafka consumer")
		return
	}

	topic := "taskTopic"

	err = consumer.Assign([]kafka.TopicPartition{
		{
			Topic:     &topic,
			Partition: 0,
			Offset:    kafka.OffsetStored,
		},
	})
	if err != nil {
		helpers.PrintErr(err, "Error assigning partitions")
		return
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigchan)
		consumer.Close()
	}()

	run := true
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Received signal: %v\n", sig)
			run = false

		default:
			ev := consumer.Poll(1)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				fmt.Printf("Received message ")

				var msg entities.TaskDta

				err := json.Unmarshal(e.Value, &msg)
				if err != nil {
					fmt.Printf("Error unmarshalling message value: %v\n", err)
					return
				}

				if err = project.Usecase.AssignTasks(msg); err != nil {
					helpers.PrintErr(err, "Error occured on AssignTasks usecase")
					continue
				}

				_, err = consumer.CommitOffsets([]kafka.TopicPartition{e.TopicPartition})
				if err != nil {
					helpers.PrintErr(err, "Error committing offset")
				}
			case (kafka.Error):
				helpers.PrintErr(e, "errror occured at consumer")
			default:
				fmt.Printf("Ignored event: %v\n", e)

			}
		}
	}

	fmt.Println("Consumer shutting down...")

}
