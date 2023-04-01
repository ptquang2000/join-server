package controllers

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/brocaar/lorawan"
	"github.com/ptquang2000/lorawan-server/models"
	"gorm.io/gorm"
)

func joinRequestHandler(msg []byte) {
	var data models.GatewayMetaData

	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		logMsg := fmt.Sprintf("Invalid message format for topic %s", joinRequestTopic)
		log.Print(logMsg)
		return
	}

	gateway := models.FindGatewayById(uint32(data.GatewayID))
	if gateway == nil {
		logMsg := fmt.Sprintf("Gateway with id %d is not registered", data.GatewayID)
		log.Print(logMsg)
		return
	}

	var phy lorawan.PHYPayload
	if err := phy.UnmarshalBinary(data.Frame); err != nil {
		logMsg := fmt.Sprintf("Error %s\n", err)
		log.Print(logMsg)
		return
	}

	jrPL, ok := phy.MACPayload.(*lorawan.JoinRequestPayload)
	if !ok {
		log.Print("MACPayload must be a *JoinRequestPayload")
		return
	}

	joinEuiBytes, _ := jrPL.JoinEUI.MarshalBinary()
	devEuiBytes, _ := jrPL.DevEUI.MarshalBinary()
	mic, _ := phy.MIC.MarshalText()

	devEui := binary.BigEndian.Uint64(devEuiBytes)
	joinEui := binary.BigEndian.Uint64(joinEuiBytes)
	devNonce := uint16(jrPL.DevNonce)

	endDevice, result := models.FindEndDeviceByDevEui(devEui)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		logMsg := fmt.Sprintf("Device with DevEui %d not found", devEui)
		log.Print(logMsg)
		return
	}

	res, err := phy.ValidateUplinkJoinMIC(lorawan.AES128Key(endDevice.Appkey))
	if !res || err != nil {
		logMsg := fmt.Sprintf("Invalid MIC in join request from %d", devEui)
		log.Print(logMsg)
		return
	}

	if endDevice.DevNonce > devNonce {
		logMsg := fmt.Sprintf("Invalid DevNonce from %d", devEui)
		log.Print(logMsg)
		return
	}

	jrFrame := models.JoinRequest{
		MacFrame: models.MacFrame{
			Major:     uint8(phy.MHDR.Major),
			FrameType: models.JOIN_REQUEST,
			Mic:       mic,
			Rssi:      data.Rssi,
			Snr:       data.Snr,
			GatewayID: data.GatewayID,
		},
		JoinEui:  joinEui,
		DevEui:   devEui,
		DevNonce: uint16(jrPL.DevNonce),
	}
	jrFrame.Create()

	if endDevice.DevNonce < devNonce {
		endDevice.DevNonce = devNonce
		joinAcceptChannel <- endDevice
	}
}

func JoinAcceptHandler(i_endDevice models.EndDevice) {
	endDevice, tx := models.FindEndDeviceById(uint32(i_endDevice.ID))
	if tx.Error != nil {
		panic("Why join accept when there is no matched end-device ?")
	}
	if i_endDevice.DevNonce <= endDevice.DevNonce {
		logMsg := fmt.Sprintf("The same join request frame with DevNonce %d might has been processed", i_endDevice.DevNonce)
		log.Print(logMsg)
		return
	}
	endDevice.DevNonce = i_endDevice.DevNonce

	frames, _ := models.FindJoinRequestByDevEuiAndDevNonce(endDevice.DevEui, endDevice.DevNonce)
	bestFrame := frames[0].MacFrame
	for _, frame := range frames[1:] {
		if !bestFrame.IsBetterGateway(frame.MacFrame) {
			bestFrame = frame.MacFrame
		}
	}

	joinNonce := endDevice.JoinNonce + 1
	jaFrame := &models.JoinAccept{
		MacFrame: models.MacFrame{
			Major:     uint8(lorawan.LoRaWANR1),
			FrameType: models.JOIN_ACCEPT,
			GatewayID: bestFrame.GatewayID,
		},
		JoinNonce: joinNonce,
		NetId:     endDevice.NetId,
		DevAddr:   endDevice.DevAddr,
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
			JoinNonce: lorawan.JoinNonce(jaFrame.JoinNonce),
			HomeNetID: lorawan.NetID(netId),
			DevAddr:   lorawan.DevAddr(devAddr),
			DLSettings: lorawan.DLSettings{
				RX2DataRate: jaFrame.RX2DataRate,
				RX1DROffset: jaFrame.RX1DROffset,
			},
			RXDelay: jaFrame.RXDelay,
		},
	}

	appKey := lorawan.AES128Key(endDevice.Appkey)
	joinEUI := make([]byte, 8)
	binary.BigEndian.PutUint64(joinEUI, endDevice.JoinEui)

	if err := phy.SetDownlinkJoinMIC(lorawan.JoinRequestType, lorawan.EUI64(joinEUI), lorawan.DevNonce(joinNonce), appKey); err != nil {
		panic(err)
	}
	if err := phy.EncryptJoinAcceptPayload(appKey); err != nil {
		panic(err)
	}
	bytes, err := phy.MarshalBinary()
	if err != nil {
		panic(err)
	}

	topic := models.FindGatewayJoinAcceptTopicById(uint32(bestFrame.GatewayID))
	if len(topic) == 0 {
		logMsg := fmt.Sprintf("Why does this gateway with id %d have no topic ?", bestFrame.GatewayID)
		panic(logMsg)
	}
	token := client.Publish(topic, 0, false, bytes)
	token.Wait()
	logMsg := fmt.Sprintf("Publish to topic %s", topic)
	log.Println(logMsg)

	mic, _ := phy.MIC.MarshalText()
	jaFrame.MacFrame.Mic = mic
	endDevice.JoinNonce = joinNonce

	jaFrame.Create()
	endDevice.Update()
}
