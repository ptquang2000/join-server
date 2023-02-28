package models

import (
	"time"
	"fmt"
	"crypto/sha256"
)

type Gateways struct {
	Id uint32 `gorm:"not null;autoIncrement"`
	Username string `gorm:"type:varchar(100);default:null;uniqueIndex:unique_username"`
	Password_harsh string `gorm:"type:varchar(100);default:null"`
	Salt string `gorm:"type:varchar(35);default:null"`
	Is_superuser bool `gorm:"default:0"`
	Created time.Time `gorm:"type:datetime;default:null"`
}

func CreateGateway() {
	username := "station1"
	password := "public"
	salt := "slat_foo123"
	is_superuser := false
	
	hash := sha256.New()
	hash.Write([]byte(password + salt))
	password_harsh := fmt.Sprintf("%x", hash.Sum(nil))

	gateway := &Gateways {
		Username: username,
		Password_harsh: password_harsh,
		Salt: salt,
		Is_superuser: is_superuser,
		Created: time.Now(),
	}
	gateway.Create()
}