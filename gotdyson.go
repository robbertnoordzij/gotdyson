package main

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const VERSION = "0.0.1"

type SensorData struct {
	Tact string `json:"tact"`
	Hact string `json:"hact"`
	Pact string `json:"pact"`
	Vact string `json:"vact"`
}

type SensorDataMessage struct {
	Msg  string     `json:"msg"`
	Time time.Time  `json:"time"`
	Data SensorData `json:"data"`
}

type stringReponse struct {
	string string
}

func (response *stringReponse) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(response.string))
}

func main() {
	var (
		listen   = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
		device   = flag.String("device-address", "", "The address of the deviceName.")
		username = flag.String("username", "", "The username for the deviceName.")
		password = flag.String("password", "", "The password for the deviceName.")
	)
	flag.Parse()

	fmt.Printf("GotDyson (%s)\n", VERSION)

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", &stringReponse{VERSION})

	sha_512 := sha512.New()
	sha_512.Write([]byte(*password))
	hashed := sha_512.Sum(nil)
	encodedPassword := base64.StdEncoding.EncodeToString(hashed)

	frame := &Frame{}
	frame.deviceName = *username

	collector := &DysonCollector{}
	prometheus.MustRegister(collector)

	opts := mqtt.NewClientOptions().AddBroker("tcp://" + *device + ":1883")
	opts.SetUsername(*username)
	opts.SetPassword(encodedPassword)
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())

		if strings.Contains(string(msg.Payload()), "HELLO") {

			splitted := strings.Split(msg.Topic(), "/")

			commandTopic := splitted[0] + "/" + splitted[1] + "/command"

			go emitCommandToDyson(client, commandTopic)
		}

		if strings.Contains(string(msg.Payload()), "ENVIRONMENTAL-CURRENT-SENSOR-DATA") {
			var sensorData SensorDataMessage
			err := json.Unmarshal(msg.Payload(), &sensorData)

			if err != nil {
				log.Printf("Could not read sensor data from %s.\n", msg.Payload())
			}

			floatValue, err := strconv.ParseFloat(sensorData.Data.Tact, 64)
			if err != nil {
				log.Printf("Could not convert sensor data from %s: %s.\n", msg.Payload(), err)
			}

			frame.temperature = floatValue / 10

			log.Printf("Received %f as temperature.\n", frame.temperature)

			collector.Update(*frame)
		}
	})
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		fmt.Printf("Connected to %s.\n", *device)
		client.Subscribe("#", 0, nil)
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	http.ListenAndServe(*listen, nil)
}

type RequestCurrentState struct {
	Msg  string    `json:"msg"`
	Time time.Time `json:"time"`
}

func emitCommandToDyson(client mqtt.Client, topic string) {

	for {
		requestCurrentState := RequestCurrentState{"REQUEST-CURRENT-STATE", time.Now()}

		jsonString, err := json.Marshal(requestCurrentState)

		if err != nil {
			fmt.Printf("Could not format request message.\n")
			continue
		}

		if token := client.Publish(topic, 1, false, jsonString); token.Wait() && token.Error() != nil {
			err = token.Error()
		}

		if err != nil {
			fmt.Printf("Could not send request message.\n")
			continue
		}

		time.Sleep(10 * time.Second)
	}
}
