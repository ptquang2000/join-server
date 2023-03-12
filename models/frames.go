package models

import (
	"gorm.io/gorm"
)

type MacFrames struct {
	gorm.Model

	FrameID uint`gorm:"uniqueIndex:unique_frame"`
	FrameType uint8 `gorm:"uniqueIndex:unique_frame"`
	Major uint8
	Payload []byte
	Mic []byte
	Rssi int8
	Snr int16
}

func ReadFrames() (frames []MacFrames) {
	return
}

type JoinRequests struct {
	gorm.Model

	MacFrame *MacFrames `gorm:"polymorphic:Frame;polymorphicValue:0"`
	
	JoinEui uint64
	DevEui uint64 `gorm:"uniqueIndex:unique_deveui"`
	DevNonce uint16
}

func (frame JoinRequests) Create() (tx *gorm.DB) {
	tx = db.Create(&frame)
	return
}

type JoinAccepts struct {
	gorm.Model

	MacFrame *MacFrames `gorm:"polymorphic:Frame;polymorphicValue:1"`

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

func (frame JoinAccepts) Create() (tx *gorm.DB) {
	tx = db.Create(&frame)
	return
}