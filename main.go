package main

import (
	"bytes"
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
	MEGAS               = 6
)

func main() {

	// chSub1 := make(chan []byte, 0)
	sub1 := func(lient MQTT.Client, msg MQTT.Message) {

		buff := bytes.NewReader(msg.Payload())
		buffer := make([]byte, 1024)

		func() {
			res := make([]byte, 0)
			defer func() {
				log.Printf("len file: %d", len(res)/1024)
			}()
			for {
				n, err := buff.Read(buffer)
				if err != nil {
					return
				}
				res = append(res, buffer[:n]...)
			}
		}()
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

	for i := 0; i < MEGAS*1024; i++ {
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
			if ok := tk.WaitTimeout(10 * time.Second); !ok {
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

	opts := MQTT.NewClientOptions().AddBroker(remoteMqttBrokerURL)
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
