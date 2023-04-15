package controllers

import (
	"crypto/aes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"

	"github.com/brocaar/lorawan"
	"github.com/pkg/errors"
	"github.com/ptquang2000/lorawan-server/models"
	"gorm.io/gorm"
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
	endDevice, result := models.FindEndDeviceByDevAddr(devAddr)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		logMsg := fmt.Sprintf("Device with DevAddr %d not found", devAddr)
		log.Print(logMsg)
		return
	}

	netId := make([]byte, 4)
	binary.BigEndian.PutUint32(netId, endDevice.NetId)
	joinEui := make([]byte, 8)
	binary.BigEndian.PutUint64(joinEui, endDevice.JoinEui)

	nwkSKey, err := getNwkSKey(
		false,
		lorawan.AES128Key(endDevice.Appkey),
		lorawan.NetID(netId),
		lorawan.EUI64(joinEui),
		lorawan.JoinNonce(endDevice.JoinNonce),
		lorawan.DevNonce(endDevice.DevNonce))

	if err != nil {
		logMsg := fmt.Sprintf("Could not get NwkSKey for %d", endDevice.DevEui)
		log.Print(logMsg)
		return
	}

	res, err := phy.ValidateUplinkDataMIC(lorawan.LoRaWAN1_0, 0, 0, 0, nwkSKey, lorawan.AES128Key{})
	if !res || err != nil {
		logMsg := fmt.Sprintf("Invalid MIC in uplink from %d", endDevice.DevEui)
		log.Print(logMsg)
		return
	}

	appSKey, err := getAppSKey(
		false,
		lorawan.AES128Key(endDevice.Appkey),
		lorawan.NetID(netId),
		lorawan.EUI64(joinEui),
		lorawan.JoinNonce(endDevice.JoinNonce),
		lorawan.DevNonce(endDevice.DevNonce))

	if err != nil {
		logMsg := fmt.Sprintf("Could not get AppSKey for %d", endDevice.DevEui)
		log.Print(logMsg)
		return
	}

	if endDevice.FCntUp != uint16(macPL.FHDR.FCnt) {
		logMsg := fmt.Sprintf("Invalid FCnt %d from %d not found", macPL.FHDR.FCnt, endDevice.DevEui)
		log.Print(logMsg)
		return
	}

	endDevice.FCntUp += 1

	// TODO: FOpts

	if *macPL.FPort == 0 {
		if err := phy.DecryptFRMPayload(nwkSKey); err != nil {
			log.Println("Decrypt FRMPayload to mac commands fail")
			return
		}
	} else {
		if err := phy.DecryptFRMPayload(appSKey); err != nil {
			log.Println("Decrypt FRMPayload to data payload fail")
			return
		}

		pl, ok := macPL.FRMPayload[0].(*lorawan.DataPayload)
		if !ok {
			log.Println("*FRMPayload must be DataPayload")
			return
		}
		fmt.Println(pl.Bytes)
	}
}

func downlinkHandler(i_endDevice models.EndDevice) {
}

func getNwkSKey(optNeg bool, nwkKey lorawan.AES128Key, netID lorawan.NetID, joinEUI lorawan.EUI64, joinNonce lorawan.JoinNonce, devNonce lorawan.DevNonce) (lorawan.AES128Key, error) {
	return getSKey(optNeg, 0x01, nwkKey, netID, joinEUI, joinNonce, devNonce)
}

func getAppSKey(optNeg bool, nwkKey lorawan.AES128Key, netID lorawan.NetID, joinEUI lorawan.EUI64, joinNonce lorawan.JoinNonce, devNonce lorawan.DevNonce) (lorawan.AES128Key, error) {
	return getSKey(optNeg, 0x02, nwkKey, netID, joinEUI, joinNonce, devNonce)
}

func getSKey(optNeg bool, typ byte, nwkKey lorawan.AES128Key, netID lorawan.NetID, joinEUI lorawan.EUI64, joinNonce lorawan.JoinNonce, devNonce lorawan.DevNonce) (lorawan.AES128Key, error) {
	var key lorawan.AES128Key
	b := make([]byte, 16)
	b[0] = typ

	netIDB, err := netID.MarshalBinary()
	if err != nil {
		return key, errors.Wrap(err, "marshal binary error")
	}

	joinEUIB, err := joinEUI.MarshalBinary()
	if err != nil {
		return key, errors.Wrap(err, "marshal binary error")
	}

	joinNonceB, err := joinNonce.MarshalBinary()
	if err != nil {
		return key, errors.Wrap(err, "marshal binary error")
	}

	devNonceB, err := devNonce.MarshalBinary()
	if err != nil {
		return key, errors.Wrap(err, "marshal binary error")
	}

	if optNeg {
		copy(b[1:4], joinNonceB)
		copy(b[4:12], joinEUIB)
		copy(b[12:14], devNonceB)
	} else {
		copy(b[1:4], joinNonceB)
		copy(b[4:7], netIDB)
		copy(b[7:9], devNonceB)
	}

	block, err := aes.NewCipher(nwkKey[:])
	if err != nil {
		return key, err
	}
	if block.BlockSize() != len(b) {
		return key, fmt.Errorf("block-size of %d bytes is expected", len(b))
	}
	block.Encrypt(key[:], b)

	return key, nil
}
