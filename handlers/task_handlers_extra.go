//task/task_handlers_extra.go

package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"myproject/database"
	"net/http"
	"time"
)

func HandleTaskID(repository *database.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
			return
		}

		task, err := repository.GetTaskByID(id)
		if err != nil {
			log.Println(err)
			http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	}
}

func HandleTaskDelete(repo *database.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, `{"error":"Идентификатор задачи не указан"}`, http.StatusBadRequest)
			return
		}

		if err := repo.DeleteTask(id); err != nil {
			log.Println(err)
			http.Error(w, `{"error":"Ошибка в базе данных"}`, http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{})
	}
}

func HandleTaskDone(repository *database.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, `{"error":"Идентификатор задачи не указан"}`, http.StatusBadRequest)
			return
		}

		task, err := repository.GetTaskByID(id)
		if err != nil {
			log.Println(err)
			http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
			return
		}

		if task.Repeat == "" {
			if err := repository.DeleteTask(id); err != nil {
				log.Println(err)
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
				return
			}
		} else {
			nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				log.Println(err)
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
				return
			}

			task.Date = nextDate
			if err := repository.UpdateTask(*task); err != nil {
				log.Println(err)
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
				return
			}
		}

		json.NewEncoder(w).Encode(map[string]interface{}{})
	}
}

func HandleTaskPut(repository *database.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var task database.Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			log.Println(err)
			http.Error(w, `{"error":"Некорректный JSON"}`, http.StatusBadRequest)
			return
		}

		if task.ID == "" {
			http.Error(w, `{"error":"Не указан идентификатор"}`, http.StatusBadRequest)
			return
		}

		if task.Date == "" || task.Title == "" {
			http.Error(w, `{"error":"Не указаны дата и заголовок"}`, http.StatusBadRequest)
			return
		}

		if err := validateDate(task.Date); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		if err := repository.UpdateTask(task); err != nil {
			log.Println(err)
			http.Error(w, `{"error":"Задача не найдена"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}
}

func validateDate(date string) error {
	parsedDate, err := time.Parse("20060102", date)
	if err != nil {
		return errors.New("Некорректный формат даты")
	}

	if parsedDate.Before(time.Now()) {
		return errors.New("Дата не может быть меньше сегодняшней")
	}

	return nil
}

func HandleTaskGet(repository *database.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		search := r.URL.Query().Get("search")
		tasks, err := repository.GetTasks(search)
		if err != nil {
			log.Println(err)
			http.Error(w, `{"error": "Ошибка получения задач"}`, http.StatusInternalServerError)
			return
		}

		if tasks == nil {
			tasks = []database.Task{}
		}

		response := map[string][]database.Task{
			"tasks": tasks,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
