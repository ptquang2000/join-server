package main

import (
	"github.com/ptquang2000/join-server/joinserver"
)

func main() {
	joinserver.DBConnect()
	joinserver.DBMigrate()
}
