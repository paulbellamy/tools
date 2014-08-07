package main

import (
	"flag"
	"fmt"
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"os"
	"time"
)

var host = flag.String("h", "localhost", "MQTT host to connect to")
var port = flag.Int("p", 1883, "MQTT port to connect to")
var username = flag.String("u", "", "MQTT Username to authenticate with.")
var password = flag.String("P", "", "MQTT Password to authenticate with.")
var topic = flag.String("t", "", "Topic to publish on")
var message = flag.String("m", "", "Message to publish")
var clientID = flag.String("i", "", "Client ID")
var uncleanSession = flag.Bool("c", false, "Clean session Flag. Include to us an unclean session.")

func CreateClient(host string, port int, clientID string, username string, password string, uncleanSession bool) *MQTT.MqttClient {
	opts := MQTT.NewClientOptions()
	opts.SetBroker(fmt.Sprintf("tcp://%s:%d", host, port))
	opts.SetTraceLevel(MQTT.Verbose)
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

func main() {
	flag.Parse()
  require("Topic", *topic)

	client := CreateClient(*host, *port, *clientID, *username, *password, *uncleanSession)

  if *message == "" {
    *message = time.Now().String()
  }

  <-client.Publish(MQTT.QOS_ONE, *topic, *message)

  client.Disconnect(250)
}
