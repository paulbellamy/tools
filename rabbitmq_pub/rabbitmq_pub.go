package main

import (
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"os"
	"runtime"
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

	if flag.Arg(0) == "" || flag.Arg(1) == "" || flag.Arg(2) == "" || flag.Arg(3) == "" {
		fmt.Printf("usage: %s [options] VHOST EXCHANGE ROUTING_KEY MESSAGE\n", os.Args[0])
    flag.PrintDefaults()
		os.Exit(1)
	}
	vhost := flag.Arg(0)
	exchange := flag.Arg(1)
	routing_key := flag.Arg(2)
	message := flag.Arg(3)

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

	err = channel.ExchangeDeclarePassive(exchange, "topic", false, true, false, false, nil)
	checkErr(err)

	err = channel.Publish(exchange, routing_key, false, false, amqp.Publishing{
		UserId: *username,
		Body:   []byte(message),
	})
	checkErr(err)

	log.Println("Publishing Message...")
}
