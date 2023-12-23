package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Task представляет собой модель задачи
type Task struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

var db *sql.DB

func init() {
	var err error
	// параметры подключения и друигие конфиг параметры использую напрямую , можно брать из конфиг файлов или переменых окружения по желанию
	db, err = sql.Open("postgres", "user=user dbname=taskdb sslmode=disable password=password")
	if err != nil {
		log.Fatal(err)
	}

	// Проверка подключения к базе данных
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Подключение к базе успешно")
	// Создание таблицы, если она не существует. Миграции и инит базы отдельно не делаю.
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255),
			status VARCHAR(50)
		);
	`)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Инициализация базы прошла успешно")
}

func main() {
	router := mux.NewRouter()

	//  Хендрелы  API
	router.HandleFunc("/tasks", getTasks).Methods("GET")
	router.HandleFunc("/tasks", createTask).Methods("POST")

	// Добавление эндпоинта для Swagger документации
	router.HandleFunc("/swagger.json", swaggerDoc).Methods("GET")

	log.Println("Сервис стартует на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// getTasks возвращает список задач
func getTasks(w http.ResponseWriter, r *http.Request) {
	status := r.FormValue("status")

	var tasks []Task

	var query string
	if status != "" {
		query = "SELECT * FROM tasks WHERE status = $1"
		log.Println("Запрос задач по статусу =", status)
	} else {
		query = "SELECT * FROM tasks"
		log.Println("Запрос задач")
	}

	rows, err := db.Query(query, status)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Status)
		if err != nil {
			log.Fatal(err)
		}
		tasks = append(tasks, task)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// createTask создает новую задачу
func createTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("Создание задачи", task)

	_, err = db.Exec("INSERT INTO tasks (title, status) VALUES ($1, $2)", task.Title, task.Status)
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusCreated)
}

// swaggerDoc возвращает Swagger документацию
func swaggerDoc(w http.ResponseWriter, r *http.Request) {
	// Простой JSON для Swagger
	swaggerJSON := `{
		"swagger": "2.0",
		"info": {
			"title": "Task API",
			"version": "1.0.0"
		},
		"paths": {
			"/tasks": {
				"get": {
					"summary": "Get tasks",
					"description": "Get a list of tasks with optional filtering by status.",
					"parameters": [
						{
							"name": "status",
							"in": "query",
							"description": "Filter tasks by status",
							"type": "string"
						}
					],
					"responses": {
						"200": {
							"description": "Successful operation",
							"schema": {
								"type": "array",
								"items": {
									"$ref": "#/definitions/Task"
								}
							}
						}
					}
				},
				"post": {
					"summary": "Create task",
					"description": "Create a new task.",
					"parameters": [
						{
							"name": "task",
							"in": "body",
							"description": "Task object",
							"required": true,
							"schema": {
								"$ref": "#/definitions/Task"
							}
						}
					],
					"responses": {
						"201": {
							"description": "Task created successfully"
						}
					}
				}
			}
		},
		"definitions": {
			"Task": {
				"type": "object",
				"properties": {
					"id": {
						"type": "integer"
					},
					"title": {
						"type": "string"
					},
					"status": {
						"type": "string"
					}
				}
			}
		}
	}`

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, swaggerJSON)
}

// Тестирование API
func TestAPI(t *testing.T) {
	// Ваш код для тестирования API здесь
}

// Тестирование хранилища данных
func TestDataStore(t *testing.T) {
	// Ваш код для тестирования хранилища данных здесь
}
