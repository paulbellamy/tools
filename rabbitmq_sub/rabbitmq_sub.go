package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/streadway/amqp"
)

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	var host = flag.String("host", "localhost", "Rabbitmq Host")
	var port = flag.Int("port", 5672, "Rabbitmq Port")
	var username = flag.String("username", "", "Username to connect with")
	var password = flag.String("password", "", "Password to connect with")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	if flag.Arg(0) == "" || flag.Arg(2) == "" {
		fmt.Printf("usage: %s [options] VHOST EXCHANGE BINDING\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	vhost := flag.Arg(0)
	exchange := flag.Arg(1)
	binding := flag.Arg(2)

	uri := amqp.URI{
		Scheme:   "amqp",
		Host:     *host,
		Port:     *port,
		Username: *username,
		Password: *password,
		Vhost:    vhost,
	}
	fmt.Println("Connecting to:", uri.String())
	connection, err := amqp.Dial(uri.String())
	checkErr(err)
	defer connection.Close()

	channel, err := connection.Channel()
	checkErr(err)
	defer channel.Close()

	queue, err := channel.QueueDeclare(
		"",    // name
		false, // durable
		true,  // auto-delete
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	checkErr(err)

	if exchange != "" {
		err = channel.ExchangeDeclarePassive(exchange, "topic", false, true, false, false, nil)
		checkErr(err)
	}

	err = channel.QueueBind(queue.Name, binding, exchange, false, nil)
	checkErr(err)

	messages, err := channel.Consume(queue.Name, "", true, true, false, false, nil)
	checkErr(err)

	log.Println("Waiting for messages...")
	for message := range messages {
		log.Printf("RoutingKey: %s\tBody: %s", message.RoutingKey, message.Body)
	}
}
