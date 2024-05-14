package service

import (
	"database/sql"
	"log"
	"os/exec"

	"commands_rest_api/model"
)

type CommandService struct {
	DB *sql.DB
}

func NewCommandService(db *sql.DB) *CommandService {
	return &CommandService{
		DB: db,
	}
}

func (cs *CommandService) CreateCommand(command *model.Command) error {
	_, err := cs.DB.Exec("INSERT INTO commands(command, result) VALUES($1, $2)", command.Command, command.Result)
	if err != nil {
		log.Printf("Error saving command output to database: %s\n", err)
		return err
	}
	return nil
}

func (cs *CommandService) GetCommandByID(id string) (*model.Command, error) {
	var command model.Command
	err := cs.DB.QueryRow("SELECT id, command, result FROM commands WHERE id = $1", id).Scan(&command.ID, &command.Command, &command.Result)
	if err != nil {
		log.Printf("Error querying command from database: %s\n", err)
		return nil, err
	}
	return &command, nil
}

func (cs *CommandService) StopCommandByID(id string) error {
	var command model.Command
	err := cs.DB.QueryRow("SELECT command FROM commands WHERE id = $1", id).Scan(&command.Command)
	if err != nil {
		log.Printf("Error querying command from database: %s\n", err)
		return err
	}

	stopCmd := exec.Command("bash", "-c", "kill -9 $(pgrep -f '"+command.Command+"')")
	if err := stopCmd.Run(); err != nil {
		log.Printf("Error stopping command: %s\n", err)
		return err
	}

	return nil
}

func (cs *CommandService) GetAllCommands() ([]model.Command, error) {
	rows, err := cs.DB.Query("SELECT id, command, result FROM commands")
	if err != nil {
		log.Printf("Error querying commands from database: %s\n", err)
		return nil, err
	}
	defer rows.Close()

	var commands []model.Command
	for rows.Next() {
		var command model.Command
		if err := rows.Scan(&command.ID, &command.Command, &command.Result); err != nil {
			log.Printf("Error scanning row: %s\n", err)
			return nil, err
		}
		commands = append(commands, command)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %s\n", err)
		return nil, err
	}

	return commands, nil
}
