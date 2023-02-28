package models

import (
	"time"
	"fmt"
	"crypto/sha256"
)

type Gateways struct {
	Id uint32 `gorm:"NOT NULL AUTO_INCREMENT"`
	Username string `gorm:"type:varchar(100) DEFAULT NULL;uniqueIndex:unique_username"`
	Password_harsh string `gorm:"type:varchar(100) DEFAULT NULL"`
	Salt string `gorm:"type:varchar(35) DEFAULT NULL"`
	Is_superuser bool `gorm:"DEFAULT 0"`
	Created time.Time `gorm:"type:datetime DEFAULT NULL"`
}

func GatewayCreate() {
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