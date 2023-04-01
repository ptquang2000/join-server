package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/brocaar/lorawan"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ptquang2000/lorawan-server/models"
)

type GatewayClient struct {
	ID          uint
	Username    string
	Password    string
	JaTopic     string
	JrTopic     string
	ConfigTopic string
}

var broker = "localhost"
var port = 1883

var joinAcceptTopic = "frames/joinaccept/%s"
var configsTopic = "gateways/%s"
var publishTopic = "frames/joinrequest"

var c = make(chan string)
var s rand.Source

const expected = 4

var count = 1

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var phy lorawan.PHYPayload
	if err := phy.UnmarshalBinary(msg.Payload()); err != nil {
		panic(err)
	}
	phyJSON, err := phy.MarshalJSON()
	if err != nil {
		panic(err)
	}
	c <- string(phyJSON)

	if count == expected {
		close(c)
	} else {
		count++
	}
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {
	gateways := []GatewayClient{
		{
			ID:          2,
			Username:    "station1",
			Password:    "123456?aD",
			JaTopic:     fmt.Sprintf(joinAcceptTopic, "station1"),
			JrTopic:     publishTopic,
			ConfigTopic: fmt.Sprintf(configsTopic, "station1"),
		},
		{
			ID:          3,
			Username:    "station2",
			Password:    "123456?aD",
			JaTopic:     fmt.Sprintf(joinAcceptTopic, "station2"),
			JrTopic:     publishTopic,
			ConfigTopic: fmt.Sprintf(configsTopic, "station2"),
		},
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	clients := []mqtt.Client{}
	for _, gw := range gateways {
		opts.SetClientID(gw.Username)
		opts.SetUsername(gw.Username)
		opts.SetPassword(gw.Password)
		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}

		fmt.Println(gw.JaTopic)
		token := client.Subscribe(gw.JaTopic, 0, messagePubHandler)
		token.Wait()
		clients = append(clients, client)
	}

	go func() {
		appKey := [16]byte{0x0e, 0xfe, 0x82, 0x00, 0x6e, 0x16, 0x80, 0xfa, 0x90, 0x05, 0x2a, 0xce, 0x4c, 0xed, 0xe3, 0x3b}
		dev_eui := [8]byte{0xAA, 0xAA, 0x0A, 0x00, 0x00, 0xFF, 0xFF, 0xFE}
		publish(clients[0], dev_eui, appKey, gateways[0])
	}()

	go func() {
		appKey := [16]byte{0x0e, 0xfe, 0x82, 0x00, 0x6e, 0x16, 0x80, 0xfa, 0x90, 0x05, 0x2a, 0xce, 0x4c, 0xed, 0xe3, 0x3b}
		dev_eui := [8]byte{0xAA, 0xAA, 0x0A, 0x00, 0x00, 0xFF, 0xFF, 0xFE}
		publish(clients[1], dev_eui, appKey, gateways[1])
	}()
	go func() {
		appKey := [16]byte{0xe9, 0x49, 0xad, 0xc4, 0xc5, 0x87, 0x72, 0x8f, 0x92, 0x60, 0x55, 0xe4, 0x6c, 0x16, 0xdc, 0xc6}
		dev_eui := [8]byte{0xBB, 0xBB, 0x0B, 0x00, 0x00, 0xFF, 0xFF, 0xFE}
		publish(clients[1], dev_eui, appKey, gateways[1])
	}()

	for res := range c {
		fmt.Println(res)
	}

	for _, client := range clients {
		client.Disconnect(250)
		fmt.Println("Client disconnected")
	}
}

func publish(client mqtt.Client, dev_eui [8]byte, appKey [16]byte, gateway GatewayClient) {
	for i := 40; i < 42; i++ {
		phy := lorawan.PHYPayload{
			MHDR: lorawan.MHDR{
				MType: lorawan.JoinRequest,
				Major: lorawan.LoRaWANR1,
			},
			MACPayload: &lorawan.JoinRequestPayload{
				JoinEUI:  [8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				DevEUI:   dev_eui,
				DevNonce: lorawan.DevNonce(i),
			},
		}

		if err := phy.SetUplinkJoinMIC(appKey); err != nil {
			panic(err)
		}

		frames, err := phy.MarshalBinary()
		if err != nil {
			panic(err)
		}

		s = rand.NewSource(time.Now().UnixNano() * int64(gateway.ID) * int64(i))
		r := rand.New(s)

		rssi := -1 * (r.Intn(120-30) + 30)
		snr := -1 * (r.Intn(20-8) + 8)

		mqtMsg := models.GatewayMetaData{
			GatewayID: gateway.ID,
			Rssi:      int8(rssi),
			Snr:       int16(snr),
			Frame:     frames,
		}

		bytes, err := json.Marshal(mqtMsg)
		if err != nil {
			panic(err)
		}

		token := client.Publish(publishTopic, 0, false, bytes)
		token.Wait()
		logMsg := fmt.Sprintf("Send message %d rssi %d, snr %d", gateway.ID, rssi, snr)
		log.Println(logMsg)
		// time.Sleep(time.Second)
	}
}
