package controllers

import (
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ptquang2000/lorawan-server/models"
)

var client mqtt.Client

var joinAcceptChannel = make(chan models.EndDevice)
var downlinkChannel = make(chan models.EndDevice)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	switch topic := msg.Topic(); topic {
	case serverConf.joinRequestTopic:
		joinRequestHandler(msg.Payload())
	case serverConf.uplinkTopic:
		uplinkHandler(msg.Payload())
	default:
		panic("Topic is not expected")
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func StartJoinServer() {
	var wg sync.WaitGroup
	wg.Add(1)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", serverConf.mqttBroker, serverConf.mqttPort))
	opts.SetClientID(serverConf.mqttUsername)
	opts.SetUsername(serverConf.mqttUsername)
	opts.SetPassword(serverConf.mqttPassword)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	topics := map[string]byte{
		serverConf.joinRequestTopic: 1,
		serverConf.uplinkTopic:      1,
	}
	token := client.SubscribeMultiple(topics, nil)
	token.Wait()

	go func() {
		for endDevice := range joinAcceptChannel {
			go func(endDevice *models.EndDevice) {
				time.Sleep(serverConf.joinAcceptDelay + serverConf.deDuplicationDelay)
				joinAcceptHandler(*endDevice)
			}(&endDevice)
		}
	}()

	go func() {
		for endDevice := range downlinkChannel {
			go func(endDevice *models.EndDevice) {
				time.Sleep(serverConf.receiveDelay + serverConf.deDuplicationDelay)
				downlinkHandler(*endDevice)
			}(&endDevice)
		}
	}()

	wg.Wait()

	client.Disconnect(250)
	fmt.Println("Client disconnected")
}
