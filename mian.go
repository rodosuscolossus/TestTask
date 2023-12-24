package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
)

// Task представляет собой модель задачи
type Task struct {
	Title  string `json:"title"`
	Status string `json:"status"`
}

var db *sql.DB

func init() {
	var err error
	// параметры подключения и друигие конфиг параметры использую напрямую , можно брать из конфиг файлов или переменых окружения по желанию
	log.Println("Начинаем подключение к базе")
	db, err = sql.Open("postgres", "user=user dbname=taskdb sslmode=disable password=password host=postgres-db-task port=5432")
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

	// Добавление эндпоинта для Swagger UI
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("#swagger-ui"),
	))

	log.Println("Сервис стартует на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// getTasks возвращает список задач
// getTasks godoc
// @Summary Get Task
// @Tags Task
// @Accept json
// @Produce json
// @Success 200 {array} Task
// @Failure 404 {object} common.Error
// @Failure 422 {object} common.Error
// @Failure 500 {object} common.Error
// @Router /tasks [get]
func getTasks(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]string

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	status := requestData["status"]

	var tasks []Task
	var query string
	var args []interface{}

	if status != "" {
		query = "SELECT * FROM tasks WHERE status = $1"
		log.Println("Запрос задач по статусу =", status)
		args = append(args, status)
	} else {
		query = "SELECT * FROM tasks"
		log.Println("Запрос задач")
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.Title, &task.Status)
		if err != nil {
			log.Fatal(err)
		}
		tasks = append(tasks, task)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// createTask создает новую задачу
// delete godoc
// @Summary createTask
// @Tags Task
// @Param Task
// @Accept json
// @Produce json
// @Success 201
// @Failure 404 {object} common.Error
// @Failure 400 {object} common.Error
// @Failure 500 {object} common.Error
// @Router /tasks [create]
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
