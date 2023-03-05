package main

import (
	"github.com/ptquang2000/join-server/models"
	"github.com/ptquang2000/join-server/controllers"
)

func main() {
	models.DBConnect()
	models.DBMigrate()

	defer models.DBClose()

	controllers.StartServer()
}
