package natsTestMessage

import (
	"log"
	"os"

	stan "github.com/nats-io/stan.go"
)

func MessageTest(sc stan.Conn) {
	jsonData, err := os.ReadFile("resources/json/model_test.json")
	if err != nil {
		log.Fatalf("Error reading JSON file: %s", err)
	}
	log.Printf("Sending message: %s", string(jsonData))

	err = sc.Publish("my_channel", jsonData)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Message published")
}
