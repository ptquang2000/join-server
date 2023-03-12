package models

import (
	"math/rand"
	"time"

	"gorm.io/gorm"
)

type EndDevices struct {
	gorm.Model

	NetId uint32 `gorm:"default:0"`
	JoinEui uint64 `gorm:"default:0"`
	DevEui uint64 `gorm:"uniqueIndex:unique_deveui" json:",string"`
	Appkey []byte `gorm:"type:blob"`
	DevAddr uint32 `gorm:"default:null;uniqueIndex:unique_devaddr"`

	JoinAccept *JoinAccepts `gorm:"foreignKey:DevAddr;references:DevAddr"`
	JoinRequest *JoinRequests `gorm:"foreignKey:DevEui;references:DevEui"`
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

func FindEndDeviceByDevAddr(devAddr uint32) (endDevice EndDevices, tx *gorm.DB) {
	tx = db.Where("dev_addr = ?", devAddr).First(&endDevice)
	return
}

func FindEndDeviceByDevEui(devEui uint64) (endDevice EndDevices, tx *gorm.DB) {
	tx = db.Where("dev_eui = ?", devEui).
		Preload("JoinRequest.MacFrame").
		Preload("JoinAccept.MacFrame").
		First(&endDevice)
	return
}

func ReadEndDevices() (endDevices []EndDevices){
	db.Find(&endDevices)
	return
}

func DeleteEndDeviceById(id uint32) (tx *gorm.DB) {
	tx = db.Delete(&EndDevices{}, id)
	return
}

func FindEndDeviceById(id uint32) (endDevice EndDevices, tx *gorm.DB) {
	tx = db.First(&endDevice, "id = ?", id)
	return
}

func (device EndDevices) Create() (tx *gorm.DB) {
	devAddr := GenerateDevAddr()
	_, result := FindEndDeviceByDevAddr(devAddr)
	for result.RowsAffected != 0 {
		devAddr = GenerateDevAddr()
		_, result = FindEndDeviceByDevAddr(devAddr)
	}
	device.DevAddr = devAddr

	tx = db.Create(&device)
	return
}

func (device EndDevices) Update() (tx *gorm.DB) {
	tx = db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&device)
	return
}