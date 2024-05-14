package main

import (
	"database/sql"
	"log"
	"net/http"
	"os/exec"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Command struct {
	ID      int    `json:"id"`
	Command string `json:"command"`
	Result  string `json:"result"`
}

var db *sql.DB

func main() {
	// Подключение к базе данных PostgreSQL
	var err error
	db, err = sql.Open("postgres", "postgres://postgres:postgres@postgres/db_commands?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Проверка подключения к базе данных
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Создание таблицы команд, если она не существует
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS commands (
        id SERIAL PRIMARY KEY,
        command TEXT NOT NULL,
        result TEXT
    )`)
	if err != nil {
		log.Fatal(err)
	}

	// Создание экземпляра маршрутизатора Gin
	r := gin.Default()

	// Отключение проксирования клиентского IP-адреса
	r.ForwardedByClientIP = false

	r.POST("/commands", createCommand)
	r.GET("/commands", getAllCommands)
	r.GET("/commands/:id", getCommandByID)

	// Запуск сервера на порту 8080
	if err := r.Run(":8000"); err != nil {
		log.Fatal(err)
	}
}

// Запускает переданную bash-команду, сохраняет результат выполнения в БД.
func createCommand(c *gin.Context) {
	var command Command
	if err := c.ShouldBindJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Received command: %s\n", command.Command)

	// Выполнение команды и получение результата
	cmd := exec.Command("bash", "-c", command.Command)
	result, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error executing command: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "output": string(result)})
		return
	}

	log.Printf("Command executed successfully, result: %s\n", result)

	// Сохранение результата выполнения команды в базу данных
	stmt, err := db.Prepare("INSERT INTO commands(command, result) VALUES($1, $2)")
	if err != nil {
		log.Printf("Error preparing SQL statement: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(command.Command, string(result))
	if err != nil {
		log.Printf("Error executing SQL statement: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Command created successfully"})
}

// Функция для обработки запроса на получение всех команд
func getAllCommands(c *gin.Context) {
	// Выборка всех команд из базы данных
	rows, err := db.Query("SELECT id, command, result FROM commands")
	if err != nil {
		log.Printf("Error querying commands from database: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var commands []Command

	// Чтение результатов запроса и добавление их в массив команд
	for rows.Next() {
		var command Command
		if err := rows.Scan(&command.ID, &command.Command, &command.Result); err != nil {
			log.Printf("Error scanning row: %s\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		commands = append(commands, command)
	}

	// Проверка наличия ошибок при чтении результатов запроса
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, commands)
}

// Функция для обработки запроса на получение одной команды по её ID
func getCommandByID(c *gin.Context) {
	// Получение ID команды из параметра маршрута
	id := c.Param("id")

	var command Command

	// Выборка команды из базы данных по её ID
	err := db.QueryRow("SELECT id, command, result FROM commands WHERE id = $1", id).Scan(&command.ID, &command.Command, &command.Result)
	if err != nil {
		log.Printf("Error querying command from database: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, command)
}
