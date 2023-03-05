package models

import (
	"gorm.io/gorm"
)

type JoinAccepts struct {
	gorm.Model

	Type uint8
	Major uint8
	JoinNonce uint32
	NetID uint32
	DevAddr uint32
	RX2DataRate uint8
	RX1DROffset uint8
	RXDelay uint8
	FreqCh3 uint32
	FreqCh4 uint32
	FreqCh5 uint32
	FreqCh6 uint32
	FreqCh7 uint32
	CFListType uint8
}

func (frame JoinAccepts) Create() {
	db.Create(&frame)
}

type JoinRequests struct {
	gorm.Model
	
	JoinEui uint64
	DevEui uint64
	DevNonce uint16
}

func (frame JoinRequests) Create() {
	db.Create(&frame)
}