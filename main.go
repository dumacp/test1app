package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	remoteMqttBrokerURL = "tcp://127.0.0.1:1883"
	// remoteMqttBrokerURL = "tcp://emqx-internal-headless:1883"
)

var megas int
var brokerURL string

func init() {
	flag.StringVar(&brokerURL, "brokerURL", "tcp://127.0.0.1:1883", "url broker")
	flag.IntVar(&megas, "megas", 1, "MiB to send")
}
func main() {

	flag.Parse()

	// chSub1 := make(chan []byte, 0)
	sub1 := func(lient MQTT.Client, msg MQTT.Message) {

		go func(data []byte) {
			log.Printf("len file: %d", len(data)/1024)
		}(msg.Payload())
	}

	consumer1, err := connectMqtt("go-consumer1")
	if err != nil {
		log.Fatalln(err)
	}
	consumer1.Subscribe("TEST/SIZE", 0, sub1)

	// go func() {
	// 	for v := range chSub1 {
	// 		fmt.Printf("size in bytes of consume message: %d\n", len(v))
	// 		//time.Sleep(1 * time.Second)
	// 	}
	// }()

	chunkmsq := make([]byte, 1024)
	for i := range chunkmsq {
		chunkmsq[i] = 0x0A
	}

	testmsg := make([]byte, 0)

	for i := 0; i < megas*1024; i++ {
		testmsg = append(testmsg, chunkmsq...)
	}

	producer1, err := connectMqtt("go-producer1")
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		for {

			log.Printf("go send test msg %d", len(testmsg)/(1000*1024))
			tk := producer1.Publish("TEST/SIZE", 1, false, testmsg)
			if ok := tk.WaitTimeout(30 * time.Second); !ok {
				log.Printf("erro timeout: %s", tk.Error())
			}
			time.Sleep(10 * time.Second)
		}
	}()

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT)
	signal.Notify(finish, syscall.SIGTERM)
	<-finish

}

func connectMqtt(clientName string) (MQTT.Client, error) {

	opts := MQTT.NewClientOptions().AddBroker(brokerURL)
	opts.SetCleanSession(true)
	opts.SetClientID(clientName)
	opts.SetAutoReconnect(true)
	conn := MQTT.NewClient(opts)
	token := conn.Connect()
	if ok := token.WaitTimeout(30 * time.Second); !ok {
		return nil, fmt.Errorf("MQTT connection failed")
	}
	return conn, nil
}
