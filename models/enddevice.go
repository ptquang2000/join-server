package models

import (
	"crypto/aes"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/brocaar/lorawan"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type EndDevice struct {
	gorm.Model

	NetId   uint32 `gorm:"default:0"`
	JoinEui uint64 `gorm:"default:0"`
	DevEui  uint64 `gorm:"uniqueIndex:unique_deveui" json:",string"`
	Appkey  []byte `gorm:"type:blob"`
	DevAddr uint32 `gorm:"default:null;uniqueIndex:unique_devaddr"`

	DevNonce  uint16 `gorm:"default:0;"`
	JoinNonce uint16 `gorm:"default:0;"`

	FCntUp   uint16 `gorm:"default:0;"`
	FCntDown uint16 `gorm:"default:0;"`

	AppSKey lorawan.AES128Key `gorm:"-:all"`
	NwkSKey lorawan.AES128Key `gorm:"-:all"`
}

func GenerateAppkey() (appkey []byte) {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	for i := 0; i < 16; i++ {
		appkey = append(appkey, byte(random.Intn(256)))
	}
	return
}

func GenerateDevAddr() (devaddr uint32) {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	devaddr = uint32(random.Intn(4294967296))
	return
}

func FindEndDeviceByDevAddr(devAddr uint32) (endDevice EndDevice, tx *gorm.DB) {
	tx = db.Where("dev_addr = ?", devAddr).First(&endDevice)
	return
}

func LoadEndDeviceByDevAddr(devAddr uint32) (endDevice EndDevice, ok bool) {
	endDevice, tx := FindEndDeviceByDevAddr(devAddr)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return endDevice, false
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
	}
	endDevice.NwkSKey = nwkSKey

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
	}
	endDevice.AppSKey = appSKey

	return endDevice, true
}

func FindEndDeviceByDevEui(devEui uint64) (endDevice EndDevice, tx *gorm.DB) {
	tx = db.Where("dev_eui = ?", devEui).First(&endDevice)
	return
}

func ReadEndDevices() (endDevices []EndDevice) {
	db.Find(&endDevices)
	return
}

func DeleteEndDeviceById(id uint32) (tx *gorm.DB) {
	tx = db.Unscoped().Delete(&EndDevice{}, id)
	return
}

func FindEndDeviceById(id uint32) (endDevice EndDevice, tx *gorm.DB) {
	tx = db.First(&endDevice, "id = ?", id)
	return
}

func GetNewDevAddr() (devAddr uint32) {
	devAddr = GenerateDevAddr()
	_, result := FindEndDeviceByDevAddr(devAddr)
	for result.RowsAffected != 0 {
		devAddr = GenerateDevAddr()
		_, result = FindEndDeviceByDevAddr(devAddr)
	}
	return
}

func (device EndDevice) Create() (tx *gorm.DB) {
	tx = db.Create(&device)
	return
}

func (device EndDevice) Update() (tx *gorm.DB) {
	tx = db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&device)
	return
}

func (device EndDevice) Save() (tx *gorm.DB) {
	tx = db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&device)
	return
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
