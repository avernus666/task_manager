package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avernus666/task_manager-rework/internal/config"
)

type projectDTO struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     int64     `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func registerProjectRoutes(r *gin.Engine, db *pgxpool.Pool, cfg config.Config) {
	g := r.Group("/projects", requireAuth(cfg.JWTSecret))
	g.POST("/", func(c *gin.Context) {
		var input struct {
			Name        string `json:"name" binding:"required,max=160"`
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			abortJSON(c, http.StatusBadRequest, "name is required")
			return
		}
		input.Name = strings.TrimSpace(input.Name)
		if input.Name == "" {
			abortJSON(c, http.StatusBadRequest, "name is required")
			return
		}
		var p projectDTO
		err := db.QueryRow(c.Request.Context(), `
			INSERT INTO projects (owner_id, name, description)
			VALUES ($1, $2, $3)
			RETURNING id, name, COALESCE(description, ''), owner_id, created_at, updated_at`,
			currentClaims(c).UserID, input.Name, input.Description,
		).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			abortJSON(c, http.StatusInternalServerError, "could not create project")
			return
		}
		c.JSON(http.StatusCreated, p)
	})
	g.GET("/", func(c *gin.Context) {
		claims := currentClaims(c)
		rows, err := db.Query(c.Request.Context(), `
			SELECT id, name, COALESCE(description, ''), owner_id, created_at, updated_at
			FROM projects
			WHERE ($1 = 'admin' OR owner_id = $2)
			ORDER BY created_at DESC LIMIT 100`, claims.Role, claims.UserID)
		if err != nil {
			abortJSON(c, http.StatusInternalServerError, "could not list projects")
			return
		}
		defer rows.Close()
		projects := make([]projectDTO, 0)
		for rows.Next() {
			var p projectDTO
			if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt, &p.UpdatedAt); err != nil {
				abortJSON(c, http.StatusInternalServerError, "could not read projects")
				return
			}
			projects = append(projects, p)
		}
		c.JSON(http.StatusOK, projects)
	})
	g.GET("/:id", func(c *gin.Context) {
		p, ok := getProject(c, db, c.Param("id"))
		if !ok {
			return
		}
		c.JSON(http.StatusOK, p)
	})
	g.PUT("/:id", func(c *gin.Context) {
		var input struct {
			Name        string `json:"name" binding:"required,max=160"`
			Description string `json:"description"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			abortJSON(c, http.StatusBadRequest, "name is required")
			return
		}
		claims := currentClaims(c)
		var p projectDTO
		err := db.QueryRow(c.Request.Context(), `
			UPDATE projects SET name=$1, description=$2, updated_at=NOW()
			WHERE id=$3 AND ($4='admin' OR owner_id=$5)
			RETURNING id, name, COALESCE(description, ''), owner_id, created_at, updated_at`,
			strings.TrimSpace(input.Name), input.Description, c.Param("id"), claims.Role, claims.UserID,
		).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			abortJSON(c, http.StatusNotFound, "project not found or access denied")
			return
		}
		c.JSON(http.StatusOK, p)
	})
	g.DELETE("/:id", func(c *gin.Context) {
		claims := currentClaims(c)
		tag, err := db.Exec(c.Request.Context(), `
			DELETE FROM projects WHERE id=$1 AND ($2='admin' OR owner_id=$3)`,
			c.Param("id"), claims.Role, claims.UserID)
		if err != nil {
			abortJSON(c, http.StatusInternalServerError, "could not delete project")
			return
		}
		if tag.RowsAffected() == 0 {
			abortJSON(c, http.StatusNotFound, "project not found or access denied")
			return
		}
		c.Status(http.StatusNoContent)
	})
}

func getProject(c *gin.Context, db *pgxpool.Pool, id string) (projectDTO, bool) {
	claims := currentClaims(c)
	var p projectDTO
	err := db.QueryRow(c.Request.Context(), `
		SELECT id, name, COALESCE(description, ''), owner_id, created_at, updated_at
		FROM projects WHERE id=$1 AND ($2='admin' OR owner_id=$3)`,
		id, claims.Role, claims.UserID,
	).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		abortJSON(c, http.StatusNotFound, "project not found or access denied")
		return projectDTO{}, false
	}
	return p, true
}
