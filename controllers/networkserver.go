package controllers

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/airtime"
	"github.com/joho/godotenv"
	"github.com/ptquang2000/lorawan-server/models"
)

type ServerConfiguration struct {
	disableDutyCycle   bool
	mqttUsername       string
	mqttPassword       string
	mqttBroker         string
	mqttPort           int
	deDuplicationDelay time.Duration
	joinRequestTopic   string
	uplinkTopic        string
	receiveDelay       time.Duration
	joinAcceptDelay    time.Duration
}

var serverConf ServerConfiguration

func init() {
	var conf map[string]string
	conf, err := godotenv.Read(".conf")
	if err != nil {
		log.Fatal("Could not find configuration file")
	}

	if disable_duty_cycle, ok := conf["disable_duty_cycle"]; ok {
		serverConf.disableDutyCycle = disable_duty_cycle == "true"
	} else {
		serverConf.disableDutyCycle = false
	}
	if deduplicationDelay, ok := conf["deduplication_delay"]; ok {
		var intDeduplicatinoDelay int
		if intDeduplicatinoDelay, err = strconv.Atoi(deduplicationDelay); err != nil {
			log.Fatal("Conf deduplication delay is in wrong format")
		}
		serverConf.deDuplicationDelay = time.Microsecond * time.Duration(intDeduplicatinoDelay)
	} else {
		serverConf.deDuplicationDelay = time.Microsecond * 200
	}
	if username, ok := conf["mqtt_username"]; ok {
		serverConf.mqttUsername = username
	} else {
		log.Fatal("Require mqtt username in .conf file")
	}
	if password, ok := conf["mqtt_password"]; ok {
		serverConf.mqttPassword = password
	} else {
		log.Fatal("Require mqtt password in .conf file")
	}
	if mqttBroker, ok := conf["mqtt_broker_url"]; ok {
		serverConf.mqttBroker = mqttBroker
	} else {
		log.Fatal("Require mqtt broker url in .conf file")
	}
	if mqtt_port, ok := conf["mqtt_port"]; ok {
		if serverConf.mqttPort, err = strconv.Atoi(mqtt_port); err != nil {
			log.Fatal("Conf mqtt port is in wrong format")
		}
	} else {
		serverConf.mqttPort = 1883
	}
	if joinRequestTopic, ok := conf["join_request_topic"]; ok {
		serverConf.joinRequestTopic = joinRequestTopic
	} else {
		log.Fatal("Require join request topic in .conf file")
	}
	if uplinkTopic, ok := conf["uplink_topic"]; ok {
		serverConf.uplinkTopic = uplinkTopic
	} else {
		log.Fatal("Require uplink topic in .conf file")
	}
	if receiveDelay, ok := conf["receive_delay_duration"]; ok {
		var intReceiveDelay int
		if intReceiveDelay, err = strconv.Atoi(receiveDelay); err != nil {
			log.Fatal("Conf receive delay is in wrong format")
		}
		if intReceiveDelay != 1 && intReceiveDelay != 2 {
			log.Fatal("Conf receive delay must be 1 or 2")
		}
		serverConf.receiveDelay = time.Second * time.Duration(intReceiveDelay)
	} else {
		serverConf.receiveDelay = time.Second * 2
	}
	if joinAcceptDelay, ok := conf["join_accept_delay_duration"]; ok {
		var intJoinAcceptDelay int
		if intJoinAcceptDelay, err = strconv.Atoi(joinAcceptDelay); err != nil {
			log.Fatal("Conf join accept delay is in wrong format")
		}
		if intJoinAcceptDelay != 5 && intJoinAcceptDelay != 6 {
			log.Fatal("Conf join accept delay must be 5 or 6")
		}
		serverConf.joinAcceptDelay = time.Second * time.Duration(intJoinAcceptDelay)
	} else {
		serverConf.joinAcceptDelay = time.Second * 6
	}
}

func uplinkHandler(msg []byte) {
	var data models.GatewayMetaData

	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		logMsg := fmt.Sprintf("Invalid message format for topic %s", serverConf.joinRequestTopic)
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

	gatewayActivities := models.GatewayActivity{
		GatewayID: gateway.ID,
		FType:     translateFrameType(phy.MHDR.MType),
		Rssi:      data.Rssi,
		Snr:       data.Snr,
	}
	gatewayActivities.Save()

	gwliveData := GatewayLiveData{
		ID:    uint64(gatewayActivities.GatewayID),
		FType: gatewayActivities.FType,
		Time:  gatewayActivities.CreatedAt,
		Rssi:  gatewayActivities.Rssi,
		Snr:   gatewayActivities.Snr,
	}
	if dataChan, required := gwLiveDataChans[uint64(gatewayActivities.GatewayID)]; required {
		dataChan <- gwliveData
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

	existedFrames := models.FindMacPayloadByMic(mic)
	if len(existedFrames) > 0 {
		macFrame.Create()
		return
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
		logMsg := fmt.Sprintf("Payload from %d with fport %d: %v (\"%s\")", endDevice.DevEui, *macPL.FPort, pl.Bytes, string(pl.Bytes))
		log.Println(logMsg)

		endDeviceActivities := models.EndDeviceActivity{
			EndDeviceID: endDevice.ID,
			FType:       macFrame.FrameType,
			Payload:     pl.Bytes,
		}
		endDeviceActivities.Save()

		liveData := EndDeviceLiveData{
			ID:      uint64(endDeviceActivities.EndDeviceID),
			FType:   endDeviceActivities.FType,
			Time:    endDeviceActivities.CreatedAt,
			Payload: endDeviceActivities.Payload,
		}
		if dataChan, required := edLiveDataChans[uint64(endDevice.ID)]; required {
			dataChan <- liveData
		}

		topic := fmt.Sprintf("%d/fport/%d", endDevice.DevEui, *macPL.FPort)
		token := client.Publish(topic, 0, true, pl.Bytes)
		token.Wait()
		logMsg = fmt.Sprintf("Publish data to topic %s", topic)
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

	frames, _ := models.FindMacFrameByDevAddrAndFCntAndTxAvailable(endDevice.DevAddr, endDevice.FCntUp-1, serverConf.disableDutyCycle)
	if len(frames) == 0 {
		log.Print("There are no gateways in off duty cycle")
		return
	}

	bestFrame := frames[0].MacFrame
	for _, frame := range frames[1:] {
		if !bestFrame.IsBetterGateway(frame.MacFrame) {
			bestFrame = frame.MacFrame
		}
	}
	bestGateway := models.FindGatewayById(uint32(bestFrame.GatewayID))

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

	// Use hardcode setting for now
	timeOnAir, err := airtime.CalculateLoRaAirtime(len(bytes), 11, 125, 8, airtime.CodingRate45, true, false)
	if err != nil {
		bestGateway.TxAvailableAt = time.Now().Add(time.Minute)
	} else {
		bestGateway.TxAvailableAt = time.Now().Add(timeOnAir * 90)
	}
	bestGateway.Save()

	mic, _ := phy.MIC.MarshalText()
	unDlFrame.MacFrame.Mic = mic
	unDlFrame.Create()

	endDevice.FCntDown += 1
	endDevice.Update()

	gatewayActivities := models.GatewayActivity{
		GatewayID: bestFrame.GatewayID,
		FType:     models.UNCONFIRMED_DATA_DOWNLINK,
	}
	gatewayActivities.Save()

	gwliveData := GatewayLiveData{
		ID:    uint64(gatewayActivities.GatewayID),
		FType: gatewayActivities.FType,
		Time:  gatewayActivities.CreatedAt,
	}
	if dataChan, required := gwLiveDataChans[uint64(gatewayActivities.GatewayID)]; required {
		dataChan <- gwliveData
	}

	endDeviceActivities := models.EndDeviceActivity{
		EndDeviceID: endDevice.ID,
		FType:       models.UNCONFIRMED_DATA_DOWNLINK,
		Payload:     []byte{},
	}
	endDeviceActivities.Save()

	liveData := EndDeviceLiveData{
		ID:      uint64(endDeviceActivities.EndDeviceID),
		FType:   endDeviceActivities.FType,
		Time:    endDeviceActivities.CreatedAt,
		Payload: endDeviceActivities.Payload,
	}
	if dataChan, required := edLiveDataChans[uint64(endDevice.ID)]; required {
		dataChan <- liveData
	}
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
