package models

import (
	"gorm.io/gorm"
)

type MacFrames struct {
	gorm.Model

	Type uint8
	Major uint8
	Payload []byte
	Mic []byte
	Rssi int8
	Snr int16
}

type JoinAccepts struct {
	gorm.Model
	MacFrame MacFrames `gorm:"foreignKey:ID"`

	JoinNonce uint16 `gorm:"default:0"`

	NetId uint32
	DevAddr uint32 `gorm:"uniqueIndex:unique_devaddr"`
	RX2DataRate uint8 `gorm:"default:0"`
	RX1DROffset uint8 `gorm:"default:0"`
	RXDelay uint8 `gorm:"default:5"`
	FreqCh3 uint32 `gorm:"default:0"`
	FreqCh4 uint32 `gorm:"default:0"`
	FreqCh5 uint32 `gorm:"default:0"`
	FreqCh6 uint32 `gorm:"default:0"`
	FreqCh7 uint32 `gorm:"default:0"`
	CFListType uint8 `gorm:"default:1"`
}

func (frame JoinAccepts) Create() {
	db.Create(&frame)
}

type JoinRequests struct {
	gorm.Model
	MacFrame MacFrames `gorm:"foreignKey:ID"`
	
	JoinEui uint64
	DevEui uint64 `gorm:"uniqueIndex:unique_deveui"`
	DevNonce uint16
}

func (frame JoinRequests) Create() (tx *gorm.DB) {
	tx = db.Create(&frame)
	return
}

func (frame JoinRequests) Save() (tx *gorm.DB) {
	tx = db.Save(&frame)
	return
}