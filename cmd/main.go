package main

import (
	"bufio"
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
	r.POST("/commands/:id/stop", stopCommandByID)

	// Запуск сервера на порту 8080
	if err := r.Run(":8000"); err != nil {
		log.Fatal(err)
	}
}

// Функция для выполнения команды асинхронно и сохранения вывода в БД
func executeAndSaveCommand(command string, done chan<- struct{}) {
	// Выполнение команды
	cmd := exec.Command("bash", "-c", command)

	// Получение потока вывода команды
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error getting stdout pipe: %s\n", err)
		return
	}

	// Запуск команды
	if err := cmd.Start(); err != nil {
		log.Printf("Error starting command: %s\n", err)
		return
	}

	// Чтение и сохранение вывода команды в БД по мере выполнения
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		output := scanner.Text()
		// Сохранение вывода в БД
		if _, err := db.Exec("INSERT INTO commands(command, result) VALUES($1, $2)", command, output); err != nil {
			log.Printf("Error saving command output to database: %s\n", err)
			// Можно добавить логику обработки ошибки, если требуется
		}
	}

	// Ожидание завершения выполнения команды
	if err := cmd.Wait(); err != nil {
		log.Printf("Command finished with error: %s\n", err)
		// Можно добавить логику обработки ошибки, если требуется
	}

	// Отправляем сигнал об окончании выполнения команды через канал
	done <- struct{}{}
}

// Функция для обработки запроса на выполнение команды
func createCommand(c *gin.Context) {
	var command Command
	if err := c.ShouldBindJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Received command: %s\n", command.Command)

	// Канал для сигнала о завершении выполнения команды
	done := make(chan struct{})

	// Запуск выполнения команды в отдельной горутине
	go executeAndSaveCommand(command.Command, done)

	// Ожидание завершения выполнения команды
	<-done

	c.JSON(http.StatusCreated, gin.H{"message": "Command execution completed"})
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

// Функция для остановки выполнения команды по её ID
func stopCommandByID(c *gin.Context) {
	// Получение ID команды из параметра маршрута
	id := c.Param("id")

	// Проверка, существует ли команда с указанным ID
	var command Command
	err := db.QueryRow("SELECT command FROM commands WHERE id = $1", id).Scan(&command.Command)
	if err != nil {
		log.Printf("Error querying command from database: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Выполнение команды остановки
	stopCmd := exec.Command("bash", "-c", "kill -9 $(pgrep -f '"+command.Command+"')")
	if err := stopCmd.Run(); err != nil {
		log.Printf("Error stopping command: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, gin.H{"message": "Command stopped successfully"})
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
