package main

import (
    "fmt"
    "time"

    "github.com/brocaar/lorawan"
    mqtt "github.com/eclipse/paho.mqtt.golang"
)

var gatewayId = "station1" 
var gatewayPassword = "station1"
var broker = "localhost"
var port = 1883

var joinAcceptTopic = "frames/joinaccept"
var configsTopic = fmt.Sprintf("gateways/%s", gatewayId)

var publishTopic = "frames/joinrequest"

var c = make(chan string)
const expected = 9
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
    opts := mqtt.NewClientOptions()
    opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
    opts.SetClientID(gatewayId)
    opts.SetUsername(gatewayId)
    opts.SetPassword(gatewayPassword)
    opts.SetDefaultPublishHandler(messagePubHandler)
    opts.OnConnect = connectHandler
    opts.OnConnectionLost = connectLostHandler
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }

    topics := map[string]byte {
        joinAcceptTopic: 1, 
        configsTopic: 1,
    }
    token := client.SubscribeMultiple(topics, nil)
    token.Wait()
    
    go func() {
        dev_eui := [8]byte{0xAA, 0xAA, 0x0A, 0x00, 0x00, 0xFF, 0xFF, 0xFE}
        publish(client, dev_eui)
    }()
    go func() {
        dev_eui := [8]byte{0xBB, 0xBB, 0x0B, 0x00, 0x00, 0xFF, 0xFF, 0xFE}
        publish(client, dev_eui)
    }()
    go func() {
        dev_eui := [8]byte{0xCC, 0xCC, 0x0C, 0x00, 0x00, 0xFF, 0xFF, 0xFE}
        publish(client, dev_eui)
    }()

    for res := range c {
        fmt.Println(res)
    }

    client.Disconnect(250)
    fmt.Println("Client disconnected")
}

func publish(client mqtt.Client, dev_eui [8]byte) {
    appKey := [16]byte{0x69, 0x6d, 0xab, 0xcf, 0x83, 0x55, 0xa5, 0x59, 0xcd, 0xed, 0x8b, 0xd3, 0xf3, 0x65, 0x57, 0xb5 }
    for i := 0; i < 3; i++ {
        phy := lorawan.PHYPayload{
            MHDR: lorawan.MHDR{
                MType: lorawan.JoinRequest,
                Major: lorawan.LoRaWANR1,
            },
            MACPayload: &lorawan.JoinRequestPayload{
                JoinEUI: [8]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
                DevEUI: dev_eui,
                DevNonce: lorawan.DevNonce(i),
            },
        }

        if err := phy.SetUplinkJoinMIC(appKey); err != nil {
            panic(err)
        }

        bytes, err := phy.MarshalBinary()
        if err != nil {
            panic(err)
        }

        token := client.Publish(publishTopic, 0, false, bytes)
        token.Wait()
        time.Sleep(time.Second)
    }
}