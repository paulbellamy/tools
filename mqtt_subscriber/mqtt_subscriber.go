package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	MQTT "github.com/errordeveloper/go-mqtt"
)

var host = flag.String("h", "localhost", "MQTT host to connect to")
var port = flag.Int("p", 1883, "MQTT port to connect to")
var username = flag.String("u", "", "MQTT Username to authenticate with.")
var password = flag.String("P", "", "MQTT Password to authenticate with.")
var topic = flag.String("t", "", "Topics to subscribe to (comma-separated)")
var clientID = flag.String("i", "", "Client ID")
var uncleanSession = flag.Bool("c", false, "Unclean session Flag. Include it to use an unclean session.")

func CreateClient(host string, port int, clientID string, username string, password string, uncleanSession bool) *MQTT.MqttClient {
	opts := MQTT.NewClientOptions()
	opts.SetBroker(fmt.Sprintf("tcp://%s:%d", host, port))
	opts.SetTraceLevel(MQTT.Warn)
	opts.SetClientId(clientID)
	opts.SetTimeout(3600)
	if username != "" {
		opts.SetUsername(username)
	}
	if password != "" {
		opts.SetPassword(password)
	}
	opts.SetCleanSession(!uncleanSession)

	return MQTT.NewClient(opts)
}

func require(name, value string) {
	if value == "" {
		fmt.Println(name, "is required")
		os.Exit(1)
	}
}

func callback(client *MQTT.MqttClient, message MQTT.Message) {
	dup := 0
	if message.DupFlag() {
		dup = 1
	}

	retained := 0
	if message.RetainedFlag() {
		retained = 1
	}

	fmt.Printf(
		"(d%d, q%d, r%d, m%d, '%s', ... (%d bytes))\n%s\n\n",
		dup,
		message.QoS(),
		retained,
		message.MsgId(),
		message.Topic(),
		len(message.Payload()),
		message.Payload(),
	)
}

func main() {
	flag.Parse()
	require("At least one topic", *topic)

	client := CreateClient(*host, *port, *clientID, *username, *password, *uncleanSession)
	_, err := client.Start()
	if err != nil {
		log.Fatal(err)
	}

	topicFilters := []*MQTT.TopicFilter{}
	for _, topic := range strings.Split(*topic, ",") {
		trimmed := strings.TrimSpace(topic)
		if trimmed == "" {
			log.Fatal("Topic cannot be empty")
		}

		filter, err := MQTT.NewTopicFilter(trimmed, 0)
		if err != nil {
			log.Fatal(err)
		}

		topicFilters = append(topicFilters, filter)
	}

	receipt, err := client.StartSubscription(callback, topicFilters...)
	if err != nil {
		log.Fatal(err)
	}
	<-receipt

	for client.IsConnected() {
		<-time.After(1 * time.Second)
	}
}
