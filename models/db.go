package models

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var newLogger = logger.New(
	log.New(os.Stdout, "\r\n", log.LstdFlags),
	logger.Config{
	  SlowThreshold: time.Second,
	  LogLevel: logger.Info,
	  IgnoreRecordNotFoundError: true, 
	  Colorful: true,
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

func DBMigrate() {
	db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4").AutoMigrate(&Gateways{})
}

type DBModels interface {
	Create()
}

func (gateway Gateways) Create() {
	db.Create(&gateway)
}