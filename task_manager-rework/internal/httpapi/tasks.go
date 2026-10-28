package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avernus666/task_manager-rework/internal/config"
)

type taskDTO struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	ProjectID   int64      `json:"project_id"`
	CreatorID   int64      `json:"creator_id"`
	AssigneeID  *int64     `json:"assignee_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func registerTaskRoutes(r *gin.Engine, db *pgxpool.Pool, cfg config.Config) {
	g := r.Group("/tasks", requireAuth(cfg.JWTSecret))
	g.POST("/", func(c *gin.Context) {
		var in struct {
			Title       string     `json:"title" binding:"required,max=200"`
			Description string     `json:"description"`
			Status      string     `json:"status"`
			Priority    string     `json:"priority"`
			Deadline    *time.Time `json:"deadline"`
			ProjectID   int64      `json:"project_id" binding:"required,gt=0"`
			AssigneeID  *int64     `json:"assignee_id"`
		}
		if err := c.ShouldBindJSON(&in); err != nil {
			abortJSON(c, http.StatusBadRequest, "invalid task data")
			return
		}
		in.Title = strings.TrimSpace(in.Title)
		if in.Title == "" {
			abortJSON(c, http.StatusBadRequest, "title is required")
			return
		}
		if !validAssignee(c, db, in.AssigneeID) {
			return
		}
		if in.Status == "" {
			in.Status = "new"
		}
		if in.Priority == "" {
			in.Priority = "medium"
		}
		if !validStatus(in.Status) || !validPriority(in.Priority) {
			abortJSON(c, http.StatusBadRequest, "invalid status or priority")
			return
		}
		project, ok := getProject(c, db, intString(in.ProjectID))
		if !ok {
			return
		}
		claims := currentClaims(c)
		var t taskDTO
		err := db.QueryRow(c.Request.Context(), `
			INSERT INTO tasks (title, description, status, priority, deadline, project_id, creator_id, assignee_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			RETURNING id,title,COALESCE(description,''),status,priority,deadline,project_id,creator_id,assignee_id,created_at,updated_at`,
			strings.TrimSpace(in.Title), in.Description, in.Status, in.Priority, in.Deadline, in.ProjectID, claims.UserID, in.AssigneeID,
		).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Deadline, &t.ProjectID, &t.CreatorID, &t.AssigneeID, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			abortJSON(c, http.StatusInternalServerError, "could not create task")
			return
		}
		sendPreferenceEmail(db, project.OwnerID, "task", "Task Manager — новая задача", "В проекте «"+project.Name+"» создана задача: "+t.Title)
		c.JSON(http.StatusCreated, t)
	})
	g.GET("/", func(c *gin.Context) {
		claims := currentClaims(c)
		if value := c.Query("status"); value != "" && !validStatus(value) {
			abortJSON(c, http.StatusBadRequest, "invalid status filter")
			return
		}
		if value := c.Query("priority"); value != "" && !validPriority(value) {
			abortJSON(c, http.StatusBadRequest, "invalid priority filter")
			return
		}
		query := `
			SELECT t.id,t.title,COALESCE(t.description,''),t.status,t.priority,t.deadline,t.project_id,t.creator_id,t.assignee_id,t.created_at,t.updated_at
			FROM tasks t JOIN projects p ON p.id=t.project_id
			WHERE ($1='admin' OR p.owner_id=$2)
			  AND ($3::bigint IS NULL OR t.project_id=$3)
			  AND ($4::text IS NULL OR t.status=$4)
			  AND ($5::text IS NULL OR t.priority=$5)
			  AND ($6::bigint IS NULL OR t.assignee_id=$6)
			ORDER BY t.created_at DESC LIMIT 100`
		rows, err := db.Query(c.Request.Context(), query, claims.Role, claims.UserID,
			nullableInt(c.Query("project_id")), nullableString(c.Query("status")),
			nullableString(c.Query("priority")), nullableInt(c.Query("assignee_id")))
		if err != nil {
			abortJSON(c, http.StatusBadRequest, "could not list tasks; check filter values")
			return
		}
		defer rows.Close()
		out := make([]taskDTO, 0)
		for rows.Next() {
			var t taskDTO
			if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Deadline, &t.ProjectID, &t.CreatorID, &t.AssigneeID, &t.CreatedAt, &t.UpdatedAt); err != nil {
				abortJSON(c, http.StatusInternalServerError, "could not read tasks")
				return
			}
			out = append(out, t)
		}
		c.JSON(http.StatusOK, out)
	})
	g.GET("/:id", func(c *gin.Context) {
		t, ok := getTask(c, db, c.Param("id"))
		if !ok {
			return
		}
		c.JSON(http.StatusOK, t)
	})
	g.PUT("/:id", func(c *gin.Context) {
		var in struct {
			Title       string     `json:"title" binding:"required,max=200"`
			Description string     `json:"description"`
			Status      string     `json:"status" binding:"required"`
			Priority    string     `json:"priority" binding:"required"`
			Deadline    *time.Time `json:"deadline"`
			ProjectID   int64      `json:"project_id" binding:"required,gt=0"`
			AssigneeID  *int64     `json:"assignee_id"`
		}
		if err := c.ShouldBindJSON(&in); err != nil || !validStatus(in.Status) || !validPriority(in.Priority) {
			abortJSON(c, http.StatusBadRequest, "invalid task data")
			return
		}
		in.Title = strings.TrimSpace(in.Title)
		if in.Title == "" {
			abortJSON(c, http.StatusBadRequest, "title is required")
			return
		}
		if !validAssignee(c, db, in.AssigneeID) {
			return
		}
		if _, ok := getProject(c, db, intString(in.ProjectID)); !ok {
			return
		}
		previous, exists := getTask(c, db, c.Param("id"))
		if !exists {
			return
		}
		claims := currentClaims(c)
		var t taskDTO
		err := db.QueryRow(c.Request.Context(), `
			UPDATE tasks SET title=$1,description=$2,status=$3,priority=$4,deadline=$5,project_id=$6,assignee_id=$7,updated_at=NOW()
			WHERE id=$8 AND EXISTS (SELECT 1 FROM projects p WHERE p.id=tasks.project_id AND ($9='admin' OR p.owner_id=$10))
			RETURNING id,title,COALESCE(description,''),status,priority,deadline,project_id,creator_id,assignee_id,created_at,updated_at`,
			strings.TrimSpace(in.Title), in.Description, in.Status, in.Priority, in.Deadline, in.ProjectID, in.AssigneeID, c.Param("id"), claims.Role, claims.UserID,
		).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Deadline, &t.ProjectID, &t.CreatorID, &t.AssigneeID, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			abortJSON(c, http.StatusNotFound, "task not found or access denied")
			return
		}
		if previous.Status != t.Status {
			sendPreferenceEmail(db, previous.CreatorID, "task", "Task Manager — изменён статус задачи", "Статус задачи «"+t.Title+"» изменён на: "+t.Status)
		}
		c.JSON(http.StatusOK, t)
	})
	g.DELETE("/:id", func(c *gin.Context) {
		claims := currentClaims(c)
		tag, err := db.Exec(c.Request.Context(), `
			DELETE FROM tasks t WHERE t.id=$1 AND EXISTS (
				SELECT 1 FROM projects p WHERE p.id=t.project_id AND ($2='admin' OR p.owner_id=$3)
			)`, c.Param("id"), claims.Role, claims.UserID)
		if err != nil {
			abortJSON(c, http.StatusInternalServerError, "could not delete task")
			return
		}
		if tag.RowsAffected() == 0 {
			abortJSON(c, http.StatusNotFound, "task not found or access denied")
			return
		}
		c.Status(http.StatusNoContent)
	})
}

func getTask(c *gin.Context, db *pgxpool.Pool, id string) (taskDTO, bool) {
	claims := currentClaims(c)
	var t taskDTO
	err := db.QueryRow(c.Request.Context(), `
		SELECT t.id,t.title,COALESCE(t.description,''),t.status,t.priority,t.deadline,t.project_id,t.creator_id,t.assignee_id,t.created_at,t.updated_at
		FROM tasks t JOIN projects p ON p.id=t.project_id
		WHERE t.id=$1 AND ($2='admin' OR p.owner_id=$3)`, id, claims.Role, claims.UserID,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Deadline, &t.ProjectID, &t.CreatorID, &t.AssigneeID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		abortJSON(c, http.StatusNotFound, "task not found or access denied")
		return taskDTO{}, false
	}
	return t, true
}

func validStatus(v string) bool   { return v == "new" || v == "in_progress" || v == "completed" }
func validPriority(v string) bool { return v == "low" || v == "medium" || v == "high" }
func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
func nullableInt(v string) any {
	if v == "" {
		return nil
	}
	return v
}
func intString(v int64) string {
	// IDs are positive integers; this avoids adding another formatting dependency.
	return strconv.FormatInt(v, 10)
}

func validAssignee(c *gin.Context, db *pgxpool.Pool, assigneeID *int64) bool {
	if assigneeID == nil {
		return true
	}
	var active bool
	err := db.QueryRow(c.Request.Context(), `SELECT is_active FROM users WHERE id=$1`, *assigneeID).Scan(&active)
	if err != nil || !active {
		abortJSON(c, http.StatusBadRequest, "assignee must be an active user")
		return false
	}
	return true
}
