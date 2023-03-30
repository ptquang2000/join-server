package models

import (
	"gorm.io/gorm"
)

type MacFrame struct {
	gorm.Model

	FrameID   uint  `gorm:"uniqueIndex:unique_frame"`
	FrameType uint8 `gorm:"uniqueIndex:unique_frame"`
	Major     uint8
	Payload   []byte
	Mic       []byte

	GatewayID uint
	Rssi      int8
	Snr       int16
	Gateway   Gateway `gorm:"foreignKey:ID;references:GatewayID"`
}

func ReadFrames() (frames []MacFrame) {
	db.Find(&frames)
	return
}

func FindFramesWithLimit(limit uint64) (frames []MacFrame) {
	db.Order("id desc").Limit(int(limit)).Find(&frames)
	return
}

type JoinRequest struct {
	gorm.Model

	MacFrame *MacFrame `gorm:"polymorphic:Frame;polymorphicValue:0"`

	JoinEui  uint64
	DevEui   uint64 `gorm:"uniqueIndex:unique_deveui"`
	DevNonce uint16
}

func (frame JoinRequest) Create() (tx *gorm.DB) {
	tx = db.Create(&frame)
	return
}

type JoinAccept struct {
	gorm.Model

	MacFrame *MacFrame `gorm:"polymorphic:Frame;polymorphicValue:1"`

	JoinNonce uint16 `gorm:"default:0"`

	NetId       uint32
	DevAddr     uint32 `gorm:"uniqueIndex:unique_devaddr"`
	RX2DataRate uint8  `gorm:"default:0"`
	RX1DROffset uint8  `gorm:"default:0"`
	RXDelay     uint8  `gorm:"default:5"`
	FreqCh3     uint32 `gorm:"default:0"`
	FreqCh4     uint32 `gorm:"default:0"`
	FreqCh5     uint32 `gorm:"default:0"`
	FreqCh6     uint32 `gorm:"default:0"`
	FreqCh7     uint32 `gorm:"default:0"`
	CFListType  uint8  `gorm:"default:1"`
}

func (frame JoinAccept) Create() (tx *gorm.DB) {
	tx = db.Create(&frame)
	return
}
