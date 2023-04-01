package models

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type GatewayMetaData struct {
	GatewayID uint   `json:"id"`
	Rssi      int8   `json:"rssi"`
	Snr       int16  `json:"snr"`
	Frame     []byte `json:"frame"`
}

type ActionType string

const (
	PUBLISH   ActionType = "publish"
	SUBSCRIBE ActionType = "subscribe"
	ALL       ActionType = "all"
)

type PermissionType string

const (
	ALLOW PermissionType = "allow"
	DENY  PermissionType = "deny"
)

type GatewayAcl struct {
	gorm.Model

	Username   string         `gorm:"type:varchar(100)"`
	ClientId   string         `gorm:"type:varchar(100)"`
	Action     ActionType     `gorm:"type:enum('publish', 'subscribe', 'all');not null;"`
	Permission PermissionType `gorm:"type:enum('allow', 'deny');not null;"`
	Topic      string         `gorm:"type:varchar(255);not null;default:'';"`
}

type Gateway struct {
	gorm.Model

	Username      string `gorm:"type:varchar(100);default:null;uniqueIndex:unique_username"`
	Password_hash string `gorm:"type:varchar(100);default:null" json:"Password"`
	Salt          string `gorm:"type:varchar(35);default:null"`
	Is_superuser  bool   `gorm:"default:0"`

	GatewayAcls []GatewayAcl `gorm:"foreignKey:Username;references:Username"`
}

func ReadGateways() (gateways []Gateway) {
	db.Find(&gateways)
	return
}

func FindGatewayById(id uint32) (gateway *Gateway) {
	tx := db.First(&gateway, id)
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	return
}

func FindGatewayJoinAcceptTopicById(id uint32) (topic string) {
	var gateway Gateway
	result := db.Preload("GatewayAcls").First(&gateway, id)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		for _, acl := range gateway.GatewayAcls {
			if acl.Action == SUBSCRIBE && acl.Permission == ALLOW {
				topic = acl.Topic
				break
			}
		}
	}
	return
}

func DeleteGatewayById(id uint32) (tx *gorm.DB) {
	gateway := FindGatewayById(id)
	var acls []GatewayAcl
	db.Where("username = ?", gateway.Username).Find(&acls)
	for _, acl := range acls {
		db.Unscoped().Delete(&acl, acl.ID)
	}

	tx = db.Unscoped().Delete(&Gateway{}, id)
	return
}

func (gateway Gateway) Create() (tx *gorm.DB) {
	hash := sha256.New()
	hash.Write([]byte(gateway.Password_hash + gateway.Salt))
	password_hash := fmt.Sprintf("%x", hash.Sum(nil))

	gateway.Password_hash = password_hash
	gateway.GatewayAcls = []GatewayAcl{
		{
			ClientId:   gateway.Username,
			Action:     PUBLISH,
			Permission: ALLOW,
			Topic:      "frames/joinrequest",
		},
		{
			ClientId:   gateway.Username,
			Action:     SUBSCRIBE,
			Permission: ALLOW,
			Topic:      fmt.Sprintf("frames/joinaccept/%s", gateway.Username),
		},
	}

	tx = db.Create(&gateway)
	return
}
