package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ptquang2000/lorawan-server/models"
)

type FakeGateway struct {
	ID          uint
	Username    string
	Password    string
	JrTopic     string
	JaTopic     string
	UlTopic     string
	DlTopic     string
	ConfigTopic string
	JrChan      chan []byte
	UlChan      chan []byte
	MqttClient  mqtt.Client
}

const (
	broker       = "localhost"
	port         = 1883
	jaTopic      = "frames/joinaccept/%s"
	configsTopic = "gateways/%s"
	jrTopic      = "frames/joinrequest"
	ulTopic      = "frames/uplink"
	dlTopic      = "frames/downlink/%s"
)

var s = rand.NewSource(time.Now().UnixNano())

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func setupGw(devices []*FakeEndDevice) (gateways map[string]*FakeGateway) {
	gateways = map[string]*FakeGateway{
		"gateway1": {
			ID:          6,
			Username:    "gateway1",
			Password:    "123456?aD",
			JrTopic:     jrTopic,
			JaTopic:     fmt.Sprintf(jaTopic, "gateway1"),
			UlTopic:     ulTopic,
			DlTopic:     fmt.Sprintf(dlTopic, "gateway1"),
			ConfigTopic: fmt.Sprintf(configsTopic, "gateway1"),
		},
		"gateway2": {
			ID:          7,
			Username:    "gateway2",
			Password:    "123456?aD",
			JrTopic:     jrTopic,
			JaTopic:     fmt.Sprintf(jaTopic, "gateway2"),
			UlTopic:     ulTopic,
			DlTopic:     fmt.Sprintf(dlTopic, "gateway1"),
			ConfigTopic: fmt.Sprintf(configsTopic, "gateway2"),
		},
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	for _, gw := range gateways {
		opts.SetClientID(gw.Username)
		opts.SetUsername(gw.Username)
		opts.SetPassword(gw.Password)
		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}

		fmt.Println(gw.JaTopic)
		token := client.Subscribe(gw.JaTopic, 0, func(c mqtt.Client, m mqtt.Message) {
			for _, device := range devices {
				device.FrameChan <- m.Payload()
			}
		})
		token.Wait()

		gw.MqttClient = client
		gw.JrChan = make(chan []byte)
	}
	return
}

func (gateway *FakeGateway) publish(frame []byte, topic string) {
	mqtMsg := getGwMetaData(frame, gateway.ID)
	bytes, err := json.Marshal(mqtMsg)
	if err != nil {
		panic(err)
	}

	token := gateway.MqttClient.Publish(topic, 0, false, bytes)
	token.Wait()
	logMsg := fmt.Sprintf("Send message %d rssi %d, snr %d", gateway.ID, mqtMsg.Rssi, mqtMsg.Snr)
	log.Println(logMsg)
}

func getGwMetaData(frame []byte, id uint) (metaData models.GatewayMetaData) {
	s = rand.NewSource(time.Now().UnixNano() * int64(id))
	r := rand.New(s)

	rssi := -1 * (r.Intn(120-30) + 30)
	snr := -1 * (r.Intn(20-8) + 8)

	metaData = models.GatewayMetaData{
		GatewayID: id,
		Rssi:      int8(rssi),
		Snr:       int16(snr),
		Frame:     frame,
	}
	return
}
