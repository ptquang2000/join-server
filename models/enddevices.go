package models

import (
	"math/rand"
	"time"
)

type EndDevices struct {
	Id uint32 `gorm:"not null;autoIncrement"`
	Netid uint32 `gorm:"default:0"`
	JoinEui uint64 `gorm:"default:0"`
	DevEui uint64 `gorm:"uniqueIndex:unique_deveui"`
	Appkey []byte `gorm:"type:blob"`
	Devaddr uint32 `gorm:"default:null;uniqueIndex:unique_devaddr"`
	DevNonce uint16 `gorm:"default:0"`
	JoinNonce uint16 `gorm:"default:0"`
}

func GenerateAppkey() (appkey []byte) {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	for i := 0; i < 16; i++ {
		appkey = append(appkey, byte(random.Intn(256)))
	}

	return
}

func CreateEndDevice() {
	netid := uint32(0)
	join_eui := uint64(0)
	dev_eui := uint64(0xFEFFFF00000FFFFF)
	appkey := GenerateAppkey()
	
	enddevice := &EndDevices {
		Netid: netid,
		JoinEui: join_eui,
		DevEui: dev_eui,
		Appkey: appkey,
	}
	enddevice.Create()
}