package main

import (
	"fmt"
)

func runGateway(gateways map[string]*FakeGateway) {
	for _, gateway := range gateways {
		go func(gateway *FakeGateway) {
			for frame := range gateway.JrChan {
				gateway.publish(frame, gateway.JrTopic)
			}
		}(gateway)
		go func(gateway *FakeGateway) {
			for frame := range gateway.UlChan {
				gateway.publish(frame, gateway.UlTopic)
			}
		}(gateway)
	}
}

func runDevice(devices []*FakeEndDevice) {
	for _, device := range devices {
		go device.startFlow()
	}
}

func testJoin(devices []*FakeEndDevice, gateways map[string]*FakeGateway) {
	start := 1
	end := 3

	go func() {
		for i := start; i < end; i++ {
			devices[0].JaWait.Add(1)

			devices[0].devNonce += 1
			go devices[0].sendJr(gateways["gateway1"])
			go devices[0].sendJr(gateways["gateway2"])

			devices[0].JaWait.Wait()
		}
	}()

	go func(devices []*FakeEndDevice, gateways map[string]*FakeGateway) {
		for i := start; i < end; i++ {
			devices[1].JaWait.Add(1)

			devices[1].devNonce += 1
			devices[1].sendJr(gateways["gateway2"])

			devices[1].JaWait.Wait()
		}
	}(devices, gateways)
}

func main() {
	devices := setUpDev()
	gateways := setupGw(devices)

	go runGateway(gateways)
	go runDevice(devices)

	fmt.Println("Enter to test join")
	testJoin(devices, gateways)
	fmt.Scanln()

	fmt.Println("Enter to test uplink")
	fmt.Scanln()

	// func() {
	// 	publishUnUl(clients[0], fakeDevices[0], gateways["gateway1"])
	// }()

	fmt.Println("Enter to exist")
	fmt.Scanln()

	for _, gateway := range gateways {
		gateway.MqttClient.Disconnect(250)
		fmt.Println("Client disconnected")
	}
}
