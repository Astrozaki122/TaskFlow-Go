package handler

import (
	"net/http"

	"task-platform/internal/database"
	"task-platform/internal/middleware"
	"task-platform/internal/view"
)

// --------------------
// MODEL (UI ONLY)
// --------------------

type TaskView struct {
	ID    int
	Title string
}

func Dashboard(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := database.DB.Query(
		"SELECT id, title FROM tasks WHERE user_id=$1 ORDER BY id DESC",
		userID,
	)

	if err != nil {
		http.Error(w, "error loading tasks", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []TaskView

	for rows.Next() {
		var t TaskView

		if err := rows.Scan(&t.ID, &t.Title); err != nil {
			http.Error(w, "scan error", http.StatusInternalServerError)
			return
		}

		tasks = append(tasks, t)
	}

	view.Templates.ExecuteTemplate(w, "dashboard.tmpl", tasks)
}

func LoginPage(w http.ResponseWriter, r *http.Request) {
	view.Templates.ExecuteTemplate(w, "login.tmpl", nil)
}
