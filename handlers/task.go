// handlers/task.go
package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"myproject/database"
)

// Task represents a task structure
type Task struct {
	ID      int64  `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handlePostTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handlePostTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&task)
	if err != nil {
		http.Error(w, `{"error":"Ошибка десериализации JSON"}`, http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		http.Error(w, `{"error":"Не указан заголовок задачи"}`, http.StatusBadRequest)
		return
	}

	dateFormat := "20060102"
	var date time.Time
	if task.Date == "" {
		date = time.Now()
		task.Date = date.Format(dateFormat)
	} else {
		date, err = time.Parse(dateFormat, task.Date)
		if err != nil {
			http.Error(w, `{"error":"Дата представлена в формате, отличном от 20060102"}`, http.StatusBadRequest)
			return
		}
	}

	if task.Repeat != "" {
		if err := validateRepeatPattern(task.Repeat); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadRequest)
			return
		}
	}

	today := time.Now().Format(dateFormat)
	if date.Format(dateFormat) < today {
		if task.Repeat == "" {
			task.Date = today
		} else {
			nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusBadRequest)
				return
			}
			task.Date = nextDate
		}
	}

	res, err := database.DB.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
		task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		http.Error(w, `{"error":"Ошибка в базе данных"}`, http.StatusInternalServerError)
		return
	}

	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, `{"error":"Ошибка в базе данных"}`, http.StatusInternalServerError)
		return
	}

	response := map[string]string{"id": strconv.FormatInt(id, 10)}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(response)
}

func splitRepeatPattern(pattern string) []string {
	return strings.Fields(pattern)
}

func validateRepeatPattern(pattern string) error {
	parts := splitRepeatPattern(pattern)
	if len(parts) < 1 {
		return fmt.Errorf("Неподдерживаемый шаблон")
	}

	unit := parts[0]
	switch unit {
	case "d", "w", "m", "y":
		if len(parts) > 1 {
			if _, err := strconv.Atoi(parts[1]); err != nil {
				return fmt.Errorf("Invalid repeat value")
			}
		} else if unit == "w" && len(parts) == 1 {
			return fmt.Errorf("Invalid repeat value for weekly repeat")
		}
	default:
		return fmt.Errorf("Неподдерживаемый шаблон")
	}

	return nil
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	startDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("Недопустимый формат даты: %w", err)
	}

	if repeat == "" {
		return "", errors.New("повторное правило не может быть пустым")
	}

	parts := strings.Split(repeat, " ")
	if len(parts) == 0 {
		return "", errors.New("Недопустимый формат правила повтора")
	}

	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return "", errors.New("Недопустимый формат правила повтора для ежедневного правила")
		}
		days, err := strconv.Atoi(parts[1])
		if err != nil || days <= 0 || days > 400 {
			return "", errors.New("недопустимый дневной интервал")
		}
		nextDate := startDate.AddDate(0, 0, days)
		for !nextDate.After(now) {
			nextDate = nextDate.AddDate(0, 0, days)
		}
		return nextDate.Format("20060102"), nil
	case "y":
		if len(parts) != 1 {
			return "", errors.New("Недопустимый формат правила повторения для годового правила")
		}
		nextDate := startDate.AddDate(1, 0, 0)
		for !nextDate.After(now) {
			nextDate = nextDate.AddDate(1, 0, 0)
		}
		return nextDate.Format("20060102"), nil
	default:
		return "", errors.New("Неподдерживаемое правило повтора")
	}
}

func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeatStr := r.FormValue("repeat")

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "Недопустимый формат даты 'now'", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateStr, repeatStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, nextDate)
}

func NextDayCalculation() {
	now := time.Date(2024, time.January, 26, 0, 0, 0, 0, time.UTC)
	date := "20240229"
	repeat := "y"

	nextDate, err := NextDate(now, date, repeat)
	if err != nil {
		log.Fatalf("Error calculating next date: %v", err)
	}
	fmt.Printf("Next Date: %s\n", nextDate)
}
