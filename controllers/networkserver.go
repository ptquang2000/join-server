package controllers

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"

	"github.com/brocaar/lorawan"
	"github.com/ptquang2000/lorawan-server/models"
)

func uplinkHandler(msg []byte) {
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

	macPL, ok := phy.MACPayload.(*lorawan.MACPayload)
	if !ok {
		log.Print("Payload must be a *MACPayload")
		return
	}

	devAddr := binary.BigEndian.Uint32(macPL.FHDR.DevAddr[:])
	endDevice, ok := models.LoadEndDeviceByDevAddr(devAddr)
	if !ok {
		logMsg := fmt.Sprintf("Device with DevAddr %d not found", devAddr)
		log.Print(logMsg)
		return
	}

	res, err := phy.ValidateUplinkDataMIC(lorawan.LoRaWAN1_0, 0, 0, 0, endDevice.NwkSKey, lorawan.AES128Key{})
	if !res || err != nil {
		logMsg := fmt.Sprintf("Invalid MIC in uplink from %d", endDevice.DevEui)
		log.Print(logMsg)
		return
	}

	mic, _ := phy.MIC.MarshalText()
	fCtrl, _ := macPL.FHDR.FCtrl.MarshalBinary()

	macFrame := models.MacPayload{
		MacFrame: models.MacFrame{
			Major:     uint8(phy.MHDR.Major),
			FrameType: translateFrameType(phy.MHDR.MType),
			Mic:       mic,
			Rssi:      data.Rssi,
			Snr:       data.Snr,
			GatewayID: data.GatewayID,
		},
		DevEui:  endDevice.DevEui,
		DevAddr: devAddr,
		FCtrl:   fCtrl,
		FCnt:    uint16(macPL.FHDR.FCnt),
		FPort:   *macPL.FPort,
	}
	macFrame.Create()

	if endDevice.FCntUp != uint16(macPL.FHDR.FCnt) {
		logMsg := fmt.Sprintf("Invalid FCnt %d from %d not found", macPL.FHDR.FCnt, endDevice.DevEui)
		log.Print(logMsg)
		return
	}

	// TODO: FOpts

	if *macPL.FPort == 0 {
		if err := phy.DecryptFRMPayload(endDevice.NwkSKey); err != nil {
			log.Println("Decrypt FRMPayload to mac commands fail")
			return
		}
	} else {
		if err := phy.DecryptFRMPayload(endDevice.AppSKey); err != nil {
			log.Println("Decrypt FRMPayload to data payload fail")
			return
		}

		pl, ok := macPL.FRMPayload[0].(*lorawan.DataPayload)
		if !ok {
			log.Println("*FRMPayload must be DataPayload")
			return
		}
		logMsg := fmt.Sprintf("Payload from %d: %s", endDevice.DevEui, string(pl.Bytes))
		log.Println(logMsg)
	}

	endDevice.FCntUp += 1
	if phy.MHDR.MType == lorawan.ConfirmedDataUp {
		downlinkChannel <- endDevice
	}

	endDevice.Update()
}

func downlinkHandler(endDevice models.EndDevice) {
	devAddr := make([]byte, 4)
	binary.BigEndian.PutUint32(devAddr, endDevice.DevAddr)

	frames, _ := models.FindMacFrameByDevAddrAndFCnt(endDevice.DevAddr, endDevice.FCntUp-1)
	bestFrame := frames[0].MacFrame
	for _, frame := range frames[1:] {
		if !bestFrame.IsBetterGateway(frame.MacFrame) {
			bestFrame = frame.MacFrame
		}
	}

	// TODO: Just use unconfirmed for now
	fctrl := lorawan.FCtrl{
		ADR:       false,
		ADRACKReq: false,
		ACK:       true,
	}

	fhdr := lorawan.FHDR{
		DevAddr: lorawan.DevAddr(devAddr),
		FCtrl:   fctrl,
		FCnt:    uint32(endDevice.FCntDown),
	}

	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.UnconfirmedDataDown,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.MACPayload{
			FHDR: fhdr,
		},
	}

	if err := phy.EncryptFRMPayload(endDevice.AppSKey); err != nil {
		panic(err)
	}
	if err := phy.SetDownlinkDataMIC(lorawan.LoRaWAN1_0, 0, endDevice.NwkSKey); err != nil {
		panic(err)
	}
	bytes, err := phy.MarshalBinary()
	if err != nil {
		panic(err)
	}

	fCtrlBytes, _ := fctrl.MarshalBinary()
	unDlFrame := &models.MacPayload{
		MacFrame: models.MacFrame{
			Major:     uint8(lorawan.LoRaWAN1_0),
			FrameType: models.UNCONFIRMED_DATA_DOWNLINK,
			GatewayID: bestFrame.GatewayID,
		},
		DevEui:  endDevice.DevEui,
		DevAddr: endDevice.DevAddr,
		FCtrl:   fCtrlBytes,
		FCnt:    endDevice.FCntDown,
	}

	topic := models.FindGatewayDownlinkTopicById(uint32(bestFrame.GatewayID))
	if len(topic) == 0 {
		logMsg := fmt.Sprintf("Why does this gateway with id %d have no topic ?", bestFrame.GatewayID)
		panic(logMsg)
	}
	token := client.Publish(topic, 0, false, bytes)
	token.Wait()
	logMsg := fmt.Sprintf("Publish to topic %s", topic)
	log.Println(logMsg)

	mic, _ := phy.MIC.MarshalText()
	unDlFrame.MacFrame.Mic = mic
	unDlFrame.Create()

	endDevice.FCntDown += 1
	endDevice.Update()
}

func translateFrameType(mType lorawan.MType) models.FrameType {
	switch mType {
	case lorawan.ConfirmedDataUp:
		return models.CONFIRMED_DATA_UPLINK
	case lorawan.UnconfirmedDataUp:
		return models.UNCONFIRMED_DATA_UPLINK
	default:
		return models.PROPRIETARY
	}
}
