package main

import (
	"log"

	"commands_rest_api/db"
	"commands_rest_api/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := gin.Default()
	r.ForwardedByClientIP = false

	commandHandler := handler.NewCommandHandler(db)

	r.POST("/commands", commandHandler.CreateCommand)
	r.GET("/commands", commandHandler.GetAllCommands)
	r.GET("/commands/:id", commandHandler.GetCommandByID)
	r.POST("/commands/:id/stop", commandHandler.StopCommandByID)

	if err := r.Run(":8000"); err != nil {
		log.Fatal(err)
	}
}
