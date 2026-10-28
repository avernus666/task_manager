package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avernus666/task_manager-rework/internal/config"
)

type commentDTO struct {
	ID        int64     `json:"id"`
	Text      string    `json:"text"`
	TaskID    int64     `json:"task_id"`
	AuthorID  int64     `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
}

func registerCommentRoutes(r *gin.Engine, db *pgxpool.Pool, cfg config.Config) {
	g := r.Group("/comments", requireAuth(cfg.JWTSecret))
	g.POST("/", func(c *gin.Context) {
		var in struct {
			Text   string `json:"text" binding:"required,max=10000"`
			TaskID int64  `json:"task_id" binding:"required,gt=0"`
		}
		if err := c.ShouldBindJSON(&in); err != nil || strings.TrimSpace(in.Text) == "" {
			abortJSON(c, http.StatusBadRequest, "text and task_id are required")
			return
		}
		task, ok := getTask(c, db, intString(in.TaskID))
		if !ok {
			return
		}
		var out commentDTO
		err := db.QueryRow(c.Request.Context(), `
			INSERT INTO comments (text, task_id, author_id) VALUES ($1,$2,$3)
			RETURNING id,text,task_id,author_id,created_at`,
			strings.TrimSpace(in.Text), in.TaskID, currentClaims(c).UserID,
		).Scan(&out.ID, &out.Text, &out.TaskID, &out.AuthorID, &out.CreatedAt)
		if err != nil {
			abortJSON(c, http.StatusInternalServerError, "could not create comment")
			return
		}
		if task.CreatorID != currentClaims(c).UserID {
			sendPreferenceEmail(db, task.CreatorID, "comment", "Task Manager — новый комментарий", "К задаче «"+task.Title+"» добавлен комментарий: "+out.Text)
		}
		c.JSON(http.StatusCreated, out)
	})
	g.GET("/task/:task_id", func(c *gin.Context) {
		if _, ok := getTask(c, db, c.Param("task_id")); !ok {
			return
		}
		rows, err := db.Query(c.Request.Context(), `
			SELECT id,text,task_id,author_id,created_at FROM comments
			WHERE task_id=$1 ORDER BY created_at ASC LIMIT 100`, c.Param("task_id"))
		if err != nil {
			abortJSON(c, http.StatusInternalServerError, "could not list comments")
			return
		}
		defer rows.Close()
		out := make([]commentDTO, 0)
		for rows.Next() {
			var item commentDTO
			if err := rows.Scan(&item.ID, &item.Text, &item.TaskID, &item.AuthorID, &item.CreatedAt); err != nil {
				abortJSON(c, http.StatusInternalServerError, "could not read comments")
				return
			}
			out = append(out, item)
		}
		c.JSON(http.StatusOK, out)
	})
}
