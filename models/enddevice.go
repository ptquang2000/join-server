package models

import (
	"math/rand"
	"time"

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

func (device EndDevice) Create() (tx *gorm.DB) {
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

func (device EndDevice) Update() (tx *gorm.DB) {
	tx = db.Session(&gorm.Session{FullSaveAssociations: true}).Updates(&device)
	return
}
