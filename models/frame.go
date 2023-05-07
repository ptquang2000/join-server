package models

import (
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

type FrameType uint

const (
	JOIN_REQUEST              FrameType = 0
	JOIN_ACCEPT               FrameType = 1
	UNCONFIRMED_DATA_UPLINK   FrameType = 2
	UNCONFIRMED_DATA_DOWNLINK FrameType = 3
	CONFIRMED_DATA_UPLINK     FrameType = 4
	CONFIRMED_DATA_DOWNLINK   FrameType = 5
	REJOIN_REQUEST            FrameType = 6
	PROPRIETARY               FrameType = 7
)

type MacFrame struct {
	gorm.Model

	FrameType FrameType
	Major     uint8
	Mic       []byte
	Rssi      int8
	Snr       int16
	GatewayID uint
}

func FindMacPayloadByMic(mic []byte) (frames []*MacPayload) {
    db.Where("mic = ?", mic).Find(&frames)
    return
}

func FindJoinRequestsByMic(mic []byte) (frames []*JoinRequest) {
    db.Where("mic = ?", mic).Find(&frames)
    return
}

func ReadLimitFrames(limit int) (frames []MacFrame) {
    frames = []MacFrame{}
	field := "created_at, frame_type, major, mic, rssi, snr, gateway_id"
	raw := fmt.Sprintf("? UNION ALL ? UNION ALL ? ORDER BY created_at DESC LIMIT %d", limit)
	db.Raw(raw,
		db.Select(field).Model(&JoinRequest{}),
		db.Select(field).Model(&JoinAccept{}),
		db.Select(field).Model(&MacPayload{})).
		Scan(&frames)
	return
}

func FindFramesWithLimit(limit uint64) (frames []MacFrame) {
	db.Order("id desc").Limit(int(limit)).Find(&frames)
	return
}

func (frame MacFrame) IsBetterGateway(other MacFrame) (result bool) {
	result = false
	if frame.IsSame(other) &&
		frame.IsBetterRssi(other) &&
		frame.IsBetterSnr(other) {
		result = true
	}
	return
}

func (frame MacFrame) IsSame(other MacFrame) (result bool) {
	result = true
	for i, value := range frame.Mic {
		if other.Mic[i] != value {
			result = false
			break
		}
	}
	return
}

func (frame MacFrame) IsBetterRssi(other MacFrame) (result bool) {
	result = (math.Abs(float64(other.Rssi)) - math.Abs(float64(frame.Rssi))) > 0
	return
}

func (frame MacFrame) IsBetterSnr(other MacFrame) (result bool) {
	a := other.Snr - 10
	b := frame.Snr - 10
	result = (math.Abs(float64(a)) - math.Abs(float64(b))) > 0
	return
}

type JoinRequest struct {
	MacFrame
	DevNonce uint16 `gorm:"default:0;"`
	DevEui   uint64
	JoinEui  uint64
}

func FindJoinRequestByDevEuiAndDevNonce(devEui uint64, devNonce uint16) (frames []JoinRequest, tx *gorm.DB) {
	tx = db.Where("dev_eui = ? and dev_nonce = ?", devEui, devNonce).Find(&frames)
	return
}

func FindJoinRequestByDevAddrAndFCntAndTxAvailable(devEui uint64, devNonce uint16, disable_duty_cycle bool) (frames []*JoinRequest, tx *gorm.DB) {
    var foundFrames []*JoinRequest
	tx = db.Where("dev_eui = ? and dev_nonce = ?", devEui, devNonce).Find(&frames)

    if disable_duty_cycle {
        frames = append(frames, foundFrames...)
    }

    for _, frame := range foundFrames {
        gateway := FindGatewayById(uint32(frame.GatewayID))
        if gateway.TxAvailableAt.Before(time.Now()) {
            frames = append(frames, frame)
        }
    }
	return
}

func (frame JoinRequest) Save() (tx *gorm.DB) {
	tx = db.Save(&frame)
	return
}

func (frame JoinRequest) Create() (tx *gorm.DB) {
	tx = db.Create(&frame)
	return
}

type JoinAccept struct {
	MacFrame
	JoinNonce   uint16 `gorm:"default:0"`
	DevAddr     uint32
	NetId       uint32
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

func (frame JoinAccept) Save() (tx *gorm.DB) {
	tx = db.Save(&frame)
	return
}

type MacPayload struct {
	MacFrame
	DevEui  uint64
	DevAddr uint32
	FCtrl   []byte
	FCnt    uint16
	FOpts   []byte
	FPort   uint8
}

func (frame MacPayload) Create() (tx *gorm.DB) {
	tx = db.Create(&frame)
	return
}

func (frame MacPayload) Save() (tx *gorm.DB) {
	tx = db.Save(&frame)
	return
}

func FindMacFrameByDevAddrAndFCnt(devAddr uint32, fCnt uint16) (frames []MacPayload, tx *gorm.DB) {
	tx = db.Where("dev_addr = ? and f_cnt = ?", devAddr, fCnt).Find(&frames)
	return
}

func FindMacFrameByDevAddrAndFCntAndTxAvailable(devAddr uint32, fCnt uint16, disable_duty_cycle bool) (frames []*MacPayload, tx *gorm.DB) {
    var foundFrames []*MacPayload
	tx = db.Where("dev_addr = ? and f_cnt = ?", devAddr, fCnt).Find(&foundFrames)

    if disable_duty_cycle {
        frames = append(frames, foundFrames...)
    }

    for _, frame := range foundFrames {
        gateway := FindGatewayById(uint32(frame.GatewayID))
        if gateway.TxAvailableAt.Before(time.Now()) {
            frames = append(frames, frame)
        }
    }
	return
}
