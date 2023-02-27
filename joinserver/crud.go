package joinserver

import (
	"fmt"
	"log"
	"time"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var newLogger = logger.New(
	log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
	logger.Config{
	  SlowThreshold: time.Second,   // Slow SQL threshold
	  LogLevel: logger.Error, // Log level
	  IgnoreRecordNotFoundError: true,           // Ignore ErrRecordNotFound error for logger
	  Colorful: true,          // Disable color
	},
  )

var username, password string = "root", "example"
var dsn, dbname string = "localhost", "mqtt_user"
var port uint = 3306

var db *gorm.DB

func DBConnect() {
	var err error
	path := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, dsn, port, dbname)
	db, err = gorm.Open(mysql.Open(path), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		panic("Cannot connect to DB")
	}	
}

type mqtt_gateways struct {
	Id uint32 `gorm:"NOT NULL AUTO_INCREMENT"`
	Username string `gorm:"type:varchar(100) DEFAULT NULL"`
	Password_harsh string `gorm:"type:varchar(100) DEFAULT NULL"`
	Salt string `gorm:"type:varchar(35) DEFAULT NULL"`
	Is_superuser bool `gorm:"DEFAULT 0"`
	Created time.Time `gorm:"type:datetime DEFAULT NULL"`
}

func DBMigrate() {
	db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4").AutoMigrate(&mqtt_gateways{})
}