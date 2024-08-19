// database/task.go
package database

type Task struct {
	ID          string `json:"id"`
	Date        string `json:"date"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Comment     string `json:"comment,omitempty"`
	Repeat      string `json:"repeat,omitempty"`
}
