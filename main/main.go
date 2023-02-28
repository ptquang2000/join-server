package main

import (
	"github.com/ptquang2000/join-server/models"
)

func main() {
	models.DBConnect()
	models.DBMigrate()

	models.GatewayCreate()
}
