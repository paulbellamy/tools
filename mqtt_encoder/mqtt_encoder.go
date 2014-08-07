package main

import (
	"log"
	"os"

	mqtt "github.com/huin/mqtt"
)

func main() {
	msg := mqtt.Connect{
		Header: mqtt.Header{
			DupFlag:  false,
			Retain:   false,
			QosLevel: mqtt.QosLevel(0),
		},
		ProtocolName:    "MQTT",
		ProtocolVersion: 4,
		CleanSession:    true,
		KeepAliveTimer:  10,
		ClientId:        "TEST_CLIENT",
		UsernameFlag:    true,
		Username:        "username",
		PasswordFlag:    true,
		Password:        "password",
	}

	file, err := os.Create("connect.mqtt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if err := msg.Encode(file); err != nil {
		log.Fatal(err)
	}
}
