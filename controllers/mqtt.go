package controllers

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ptquang2000/lorawan-server/models"
)

var client mqtt.Client

const username = "joinserver1"
const password = "123456?aD"
const broker = "localhost"
const port = 1883

const deDuplicationDelay = 200

var joinAcceptChannel = make(chan models.EndDevice)

const joinRequestTopic = "frames/joinrequest"

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	switch topic := msg.Topic(); topic {
	case joinRequestTopic:
		go joinRequestHandler(msg.Payload())
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
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID(username)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	token := client.Subscribe(joinRequestTopic, 1, nil)
	token.Wait()

	for endDevice := range joinAcceptChannel {
		time.Sleep(time.Millisecond * deDuplicationDelay)
		JoinAcceptHandler(endDevice)
	}

	client.Disconnect(250)
	fmt.Println("Client disconnected")
}
