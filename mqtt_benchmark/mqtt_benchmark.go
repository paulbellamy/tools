package main

import (
	"encoding/csv"
	//"log"
	//"net/http"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	MQTT "github.com/errordeveloper/go-mqtt"
)

var duration = flag.Duration("duration", 180*time.Second, "Duration of the test")
var ramp = flag.Duration("ramp", 90*time.Second, "Ramp-up time, Must be < duration")
var count = flag.Int64("count", 10, "Number of clients to create")
var host = flag.String("host", "localhost", "MQTT host to connect to")
var port = flag.Int("port", 1883, "MQTT port to connect to")
var credentialsPath = flag.String("credentials", "", "CSV File to pull credentials from.")
var topic = flag.String("topic", "/voltage", "MQTT topic to subscribe/publish to")

type Result struct {
	id     int64
	action string
	start  time.Time
	finish time.Time
	err    error
	client *MQTT.MqttClient
}

var payload = "Contrary to popular belief, Lorem Ipsum is not simply random text. It has roots in a piece of classi    cal Latin literature from 45 BC, making it over 2000 years old. Richard McClintock, a Latin professor at Hampden-Sydney College in Virginia, looked up one of the more obscure Latin words    , consectetur, from a Lorem Ipsum passage, and going through the cites of the word in classical literature, discovered the undoubtable source. Lorem Ipsum comes from sections 1.10.32 and     1.10.33 of \"de Finibus Bonorum et Malorum\" (The Extremes of Good and Evil) by Cicero, written in 45 BC. This book is a treatise on the theory of ethics, very popular during the Renaissa    nce. The first line of Lorem Ipsum, \"Lorem ipsum dolor sit amet..\", comes from a line in section 1.10.32."

func CreateClient(done <-chan time.Time, id int64, username string, password string, topic string, wait time.Duration, results chan *Result) {
	select {
	case <-done:
		return
	case <-time.After(wait):
		opts := MQTT.NewClientOptions()
		opts.SetBroker(fmt.Sprintf("tcp://%s:%d", *host, *port))
		opts.SetTraceLevel(MQTT.Off)
		opts.SetClientId(fmt.Sprint(id))
		opts.SetOnConnectionLost(func(client *MQTT.MqttClient, reason error) {
			fmt.Fprintf(os.Stderr, "[error] go-mqtt suffered fatal error %v\n", reason)
		})
		opts.SetTimeout(3600)
		if username != "" {
			opts.SetUsername(username)
		}
		if password != "" {
			opts.SetPassword(password)
		}
		opts.SetCleanSession(true)

		client := MQTT.NewClient(opts)

		reportAction := func(action string, dostuff func() error) error {
			start := time.Now()
			err := dostuff()
			finish := time.Now()
			results <- &Result{
				id:     id,
				action: action,
				start:  start,
				finish: finish,
				err:    err,
				client: client,
			}
			return err
		}

		err := reportAction("connect", func() error {
			_, err := client.Start()
			return err
		})
		if err != nil {
			return
		}

		time.Sleep(10 * time.Second)

		err = reportAction("subscribe", func() error {
			topicFilter, err := MQTT.NewTopicFilter(topic, 0)
			if err != nil {
				return err
			}

			suback, err := client.StartSubscription(
				func(client *MQTT.MqttClient, message MQTT.Message) {
					// no op
				},
				topicFilter,
			)
			if err != nil {
				return err
			}

			<-suback

			return nil
		})
		if err != nil {
			return
		}

		time.Sleep(10 * time.Second)

		for i := 1; i <= 10; i++ {
			time.Sleep(5 * time.Second)
			reportAction("publish", func() error {
				<-client.Publish(MQTT.QOS_ZERO, topic, []byte(payload))
				return nil
			})
		}

		time.Sleep(800 * time.Second)

		reportAction("disconnect", func() error {
			if client.IsConnected() {
				client.Disconnect(250)
			}
			return nil
		})
	}
}

func connectedClients(clients *[]*MQTT.MqttClient) int {
	count := 0
	for i := 0; i < len(*clients); i++ {
		if (*clients)[i].IsConnected() {
			count++
		}
	}
	return count
}

func meanDuration(results []*Result) time.Duration {
	duration := 0 * time.Nanosecond
	if len(results) == 0 {
		return duration
	}

	for i := 0; i < len(results); i++ {
		duration += results[i].finish.Sub(results[i].start)
	}

	return time.Duration(duration.Nanoseconds() / int64(len(results)))
}

func outputLine(tick int, clients *[]*MQTT.MqttClient, errors []error, results map[string][]*Result) {
	fmt.Printf("%12v%12d%12d%12d%12v%12v%12v%12v\n", tick, len(*clients), connectedClients(clients), len(errors), len(results["connect"]), len(results["subscribe"]), len(results["publish"]), len(results["disconnect"]))
}

func Output(results <-chan *Result, clients *[]*MQTT.MqttClient) {
	ticker := time.NewTicker(1 * time.Second)
	iteration_counter := 0
	errors := []error{}
	collected_results := map[string][]*Result{}

	fmt.Printf("%12s%12s%12s%12s%12s%12s%12s%12s\n", "time", "clients", "connected", "errors", "connect", "subscribe", "publish", "disconnect")
	outputLine(iteration_counter, clients, errors, collected_results)
	for {
		select {
		case <-ticker.C:
			iteration_counter++
			outputLine(iteration_counter, clients, errors, collected_results)
			errors = []error{}
		case result, ok := <-results:
			if !ok {
				ticker.Stop()
				return
			}

			if _, found := collected_results[result.action]; !found {
				collected_results[result.action] = []*Result{}
			}
			collected_results[result.action] = append(collected_results[result.action], result)
			if result.action == "connect" {
				*clients = append(*clients, result.client)
			}
			if result.err != nil {
				fmt.Println("Error during", result.action, ":", result.err)
				errors = append(errors, result.err)
			}
		}
	}
}

var nextCredentialPointer = 0
var credentials [][]string

func loadCredentials(path string) error {
	credentialsFile, err := os.Open(path)
	if err != nil {
		return err
	}

	credentials, err = csv.NewReader(credentialsFile).ReadAll()
	if err != nil {
		return err
	}

	return nil
}

func nextCredentials() (string, string) {
	nextCredentialPointer++
	if nextCredentialPointer >= len(credentials) {
		nextCredentialPointer = 0
	}

	return credentials[nextCredentialPointer][0], credentials[nextCredentialPointer][1]
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	if *credentialsPath == "" {
		fmt.Println("credentials file is required")
		os.Exit(1)
	}

	if *ramp > *duration {
		fmt.Println("ramp must be > duration")
		os.Exit(1)
	}

	err := loadCredentials(*credentialsPath)
	if err != nil {
		panic(err)
	}

	done := time.After(*duration)

	results := make(chan *Result)
	clients := []*MQTT.MqttClient{}
	go Output(results, &clients)

	step_size := (*ramp).Nanoseconds() / *count
	var i int64
	for i = 0; i < *count; i++ {
		username, password := nextCredentials()
		go CreateClient(done, 1, username, password, *topic, time.Duration(i*step_size), results)
	}

	<-done

	close(results)
	for j := 0; j < len(clients); j++ {
		if clients[j].IsConnected() {
			clients[j].Disconnect(250)
		}
	}
}
