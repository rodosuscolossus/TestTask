package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Тестирование API
func TestDataStore(t *testing.T) {
	// Тест на добавление новой задачи
	t.Run("CreateTask", func(t *testing.T) {
		task := Task{Title: "New Task", Status: "Pending"}

		// Очищаем базу данных перед тестом
		_, err := db.Exec("DELETE FROM tasks")
		assert.NoError(t, err)

		// Добавляем новую задачу
		_, err = db.Exec("INSERT INTO tasks (title, status) VALUES ($1, $2)", task.Title, task.Status)
		assert.NoError(t, err)
	})

	// Тест на получение задачи
	t.Run("GetTask", func(t *testing.T) {
		// Получаем задачу из базы данных
		row := db.QueryRow("SELECT * FROM tasks LIMIT 1")

		var task Task
		err := row.Scan(&task.Title, &task.Status)
		assert.NoError(t, err)

		// Проверяем, что задача получена корректно
		assert.Equal(t, "New Task", task.Title)
		assert.Equal(t, "Pending", task.Status)
	})
}
