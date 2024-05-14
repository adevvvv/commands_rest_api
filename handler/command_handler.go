package handler

import (
	"database/sql"
	"log"
	"net/http"

	"commands_rest_api/model"
	"commands_rest_api/service"

	"github.com/gin-gonic/gin"
)

type CommandHandler struct {
	Service *service.CommandService
}

func NewCommandHandler(db *sql.DB) *CommandHandler {
	return &CommandHandler{
		Service: service.NewCommandService(db),
	}
}

func (ch *CommandHandler) CreateCommand(c *gin.Context) {
	var command model.Command
	if err := c.ShouldBindJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Received command: %s\n", command.Command)

	if err := ch.Service.CreateCommand(&command); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Command execution completed"})
}

func (ch *CommandHandler) GetCommandByID(c *gin.Context) {
	id := c.Param("id")

	command, err := ch.Service.GetCommandByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, command)
}

func (ch *CommandHandler) StopCommandByID(c *gin.Context) {
	id := c.Param("id")

	if err := ch.Service.StopCommandByID(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Command stopped successfully"})
}

func (ch *CommandHandler) GetAllCommands(c *gin.Context) {
	commands, err := ch.Service.GetAllCommands()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, commands)
}
