package servers

import (
    "fmt"
    mqtt "github.com/eclipse/paho.mqtt.golang"
    "time"
)

var username = "joinserver1" 
var password = "123456?aD"
var broker = "localhost"
var port = 1883

var joinRequestTopic = "frames/joinrequest"
var joinAcceptTopic = "frames/joinaccept"

var receivedMsgChannel = make(chan string)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
    receivedMsgChannel <- fmt.Sprintf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
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
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }

    subscribe(client)
    go publish(client)

    for {
        msg := <- receivedMsgChannel
        fmt.Println(msg)
    }

    client.Disconnect(250)
    fmt.Println("Client disconnected")
}

func publish(client mqtt.Client) {
    for i := 0; true; i++ {
        text := fmt.Sprintf("Message %d", i)
        token := client.Publish(joinAcceptTopic, 0, false, text)
        token.Wait()
        time.Sleep(time.Second)
    }
}

func subscribe(client mqtt.Client) {
    token := client.Subscribe(joinRequestTopic, 1, nil)
    token.Wait()
    fmt.Printf("Subscribed to topic: %s", joinRequestTopic)
}