package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setup()
	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil
}

func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	return declareExchange(channel)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (consumer *Consumer) Listen(topics []string) error {
	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}

	for _, s := range topics {
		err := ch.QueueBind(
			q.Name,
			s,
			"logs_topic",
			false,
			nil,
		)

		if err != nil {
			return err
		}
	}

	// messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	messages, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)
	go func() {
		// for d := range messages {
		// 	log.Println("Received:", string(d.Body))
		// 	var payload Payload
		// 	_ = json.Unmarshal(d.Body, &payload)

		// 	go handlePayload(payload)
		// }
		for d := range messages {
			log.Println("Received:", string(d.Body))

			var payload Payload

			err := json.Unmarshal(d.Body, &payload)
			if err != nil {
				log.Println("JSON ERROR:", err)
				continue
			}

			err = handlePayload(payload)
			if err != nil {
				log.Println("HANDLE ERROR:", err)
				continue
			}

			d.Ack(false)
			log.Println("Message processed successfully")
		}
	}()

	fmt.Printf("Waiting for message [Exchange, Queue] [logs_topic, %s]\n", q.Name)
	<-forever

	return nil

}

func handlePayload(payload Payload) error {
	switch payload.Name {

	case "log", "event":
		return logEvent(payload)

	case "auth":
		log.Println("auth event received")
		return nil

	default:
		return logEvent(payload)
	}
}

func logEvent(entry Payload) error {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://logger-service:8080/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return err
	}

	return nil
}
