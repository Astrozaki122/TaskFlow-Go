package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"task-platform/internal/database"
	"task-platform/internal/model"
)

func getUserID(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value("user_id").(int)
	return userID, ok
}

func CreateTask(w http.ResponseWriter, r *http.Request) {

	userID, ok := getUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var t model.Task

	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if t.Title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(
		"INSERT INTO tasks (user_id, title) VALUES ($1, $2)",
		userID,
		t.Title,
	)

	if err != nil {
		http.Error(w, "failed to create task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("task created"))
}

func GetTasks(w http.ResponseWriter, r *http.Request) {

	userID, ok := getUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := database.DB.Query(
		"SELECT id, title FROM tasks WHERE user_id=$1 ORDER BY id DESC",
		userID,
	)

	if err != nil {
		http.Error(w, "failed to fetch tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []model.Task

	for rows.Next() {
		var t model.Task

		if err := rows.Scan(&t.ID, &t.Title); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}

		t.UserID = userID
		tasks = append(tasks, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func DeleteTask(w http.ResponseWriter, r *http.Request) {

	userID, ok := getUserID(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "missing task id", http.StatusBadRequest)
		return
	}

	taskID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	result, err := database.DB.Exec(
		"DELETE FROM tasks WHERE id=$1 AND user_id=$2",
		taskID,
		userID,
	)

	if err != nil {
		http.Error(w, "failed to delete task", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	w.Write([]byte("task deleted"))
}
