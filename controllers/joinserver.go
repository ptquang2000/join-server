package controllers

import (
    "fmt"
    "errors"
    "encoding/binary"
    "log"

    "github.com/brocaar/lorawan"
    mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ptquang2000/lorawan-server/models"
	"gorm.io/gorm"
)

const username = "joinserver1" 
const password = "123456?aD"
const broker = "localhost"
const port = 1883

const joinRequestTopic = "frames/joinrequest"
const joinAcceptTopic = "frames/joinaccept"

var joinAcceptChannel = make(chan models.EndDevices)

func joinRequestHandler(frame []byte) {
    var phy lorawan.PHYPayload
	if err := phy.UnmarshalBinary(frame); err != nil {
		panic(err)
	}

	jrPL, ok := phy.MACPayload.(*lorawan.JoinRequestPayload)
	if !ok {
		panic("MACPayload must be a *JoinRequestPayload")
	}

    joinEui, _ := jrPL.JoinEUI.MarshalBinary()
    devEui, _ := jrPL.DevEUI.MarshalBinary()

    var joinRequestFrame = models.JoinRequests {
        MacFrames: models.MacFrames{
            Type: uint8(phy.MHDR.MType),
            Major: uint8(phy.MHDR.Major),
            Payload: frame[1: len(frame) - 4],
            Mic: frame[len(frame) - 4: len(frame)],
        },
        JoinEui: binary.BigEndian.Uint64(joinEui),
        DevEui:  binary.BigEndian.Uint64(devEui),
        DevNonce: uint16(jrPL.DevNonce),
    }

    endDevice, result := models.FindEndDeviceByDevEui(joinRequestFrame.DevEui)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        log.Println(fmt.Sprintf("Device with DevEui %d not found", joinRequestFrame.DevEui))
        return
    }

    if endDevice.JoinRequest == nil {
        fmt.Println(endDevice.JoinRequest)
        joinRequestFrame.Create()
        endDevice.JoinRequest = &joinRequestFrame
    } else if endDevice.JoinRequest.DevNonce != joinRequestFrame.DevNonce - 1 {
        fmt.Println("JoinRequest.DevNonce != joinRequestFrame.DevNonce - 1")
        return
    }
    endDevice.JoinRequest.DevNonce += 1
    endDevice.Update()
    joinAcceptChannel <- endDevice
}

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {

    switch topic := msg.Topic(); topic {
	case joinRequestTopic:
		joinRequestHandler(msg.Payload())
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
    client := mqtt.NewClient(opts)
    if token := client.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }

    token := client.Subscribe(joinRequestTopic, 1, nil)
    token.Wait()

    for endDevice := range joinAcceptChannel {
        var joinAcceptFrame = endDevice.JoinAccept
        
        if joinAcceptFrame == nil {
            joinAcceptFrame = &models.JoinAccepts{
                NetId: endDevice.NetId,
                DevAddr: endDevice.DevAddr,
            }
            joinAcceptFrame.Create()
        } else {
            fmt.Println(endDevice.JoinAccept)
            joinAcceptFrame.JoinNonce += 1
        }

        netId := make([]byte, 4)
        binary.BigEndian.PutUint32(netId, endDevice.NetId)

        devAddr := make([]byte, 4)
        binary.BigEndian.PutUint32(devAddr, endDevice.DevAddr)

        phy := lorawan.PHYPayload{
            MHDR: lorawan.MHDR{
                MType: lorawan.JoinAccept,
                Major: lorawan.LoRaWANR1,
            },
            MACPayload: &lorawan.JoinAcceptPayload{
                JoinNonce: lorawan.JoinNonce(joinAcceptFrame.JoinNonce),
                HomeNetID: lorawan.NetID(netId),
                DevAddr: lorawan.DevAddr(devAddr),
                DLSettings: lorawan.DLSettings {
                    RX2DataRate: joinAcceptFrame.RX2DataRate,
                    RX1DROffset: joinAcceptFrame.RX1DROffset,
                },
                RXDelay: joinAcceptFrame.RXDelay,
            },
        }

        appKey := lorawan.AES128Key(endDevice.Appkey)
        joinEUI := make([]byte, 8)
        binary.BigEndian.PutUint64(joinEUI, endDevice.JoinEui)
        joinNonce := lorawan.DevNonce(joinAcceptFrame.JoinNonce)

        if err := phy.SetDownlinkJoinMIC(lorawan.JoinRequestType, lorawan.EUI64(joinEUI), joinNonce, appKey); err != nil {
            panic(err)
        }
        if err := phy.EncryptJoinAcceptPayload(appKey); err != nil {
            panic(err)
        }
        bytes, err := phy.MarshalBinary()
        if err != nil {
            panic(err)
        }

        token := client.Publish(joinAcceptTopic, 0, false, bytes)
        token.Wait()

        payload, _ := phy.MACPayload.MarshalBinary()
        mic, _ := phy.MIC.MarshalText()
        joinAcceptFrame.MacFrames = models.MacFrames{
            Type: uint8(lorawan.JoinAccept),
            Major: uint8(lorawan.LoRaWANR1),
            Payload: payload,
            Mic: mic,
        }

        endDevice.JoinAccept = joinAcceptFrame
        endDevice.Update()
    }

    client.Disconnect(250)
    fmt.Println("Client disconnected")
}