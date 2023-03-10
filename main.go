package main

import (
	"github.com/ptquang2000/lorawan-server/models"
	"github.com/ptquang2000/lorawan-server/controllers"
	"github.com/ptquang2000/lorawan-server/servers"
)

func main() {
	models.DBConnect()
	models.DBMigrate()

	defer models.DBClose()

	go servers.StartJoinServer()

	controllers.StartServer()
}
