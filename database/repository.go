// database/repository.go
package database

import (
	"database/sql"
	"errors"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (repo *Repository) GetTaskByID(id string) (*Task, error) {
	query := "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?"
	row := repo.DB.QueryRow(query, id)

	var task Task
	if err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Задача не найдена")
		}
		return nil, err
	}

	return &task, nil
}

func (repo *Repository) UpdateTask(task Task) error {
	query := "UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?"
	result, err := repo.DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("Задача не найдена")
	}
	return nil
}

func (repo *Repository) DeleteTask(id string) error {
	query := "DELETE FROM scheduler WHERE id = ?"
	result, err := repo.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("Задача не найдена")
	}
	return nil
}

func (repo *Repository) MarkTaskAsDone(id string) error {
	query := "UPDATE scheduler SET done = 1 WHERE id = ?"
	_, err := repo.DB.Exec(query, id)
	return err
}

func (repo *Repository) GetTasks(search string) ([]Task, error) {
	query := "SELECT id, date, title, comment, repeat FROM scheduler"
	args := []interface{}{}

	if search != "" {
		query += " WHERE title LIKE ? OR comment LIKE ?"
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	query += " ORDER BY date LIMIT 50"

	rows, err := repo.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}
