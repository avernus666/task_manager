package httpapi

import (
	"context"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avernus666/task_manager-rework/internal/config"
	"github.com/avernus666/task_manager-rework/internal/notifications"
)

type notificationPreferences struct {
	Email                 string `json:"email"`
	EmailEnabled          bool   `json:"email_enabled"`
	BrowserEnabled        bool   `json:"browser_enabled"`
	TaskUpdatesEnabled    bool   `json:"task_updates_enabled"`
	CommentUpdatesEnabled bool   `json:"comment_updates_enabled"`
}

func registerNotificationRoutes(r *gin.Engine, db *pgxpool.Pool, cfg config.Config) {
	g := r.Group("/settings/notifications", requireAuth(cfg.JWTSecret))
	g.GET("/", func(c *gin.Context) {
		prefs, err := getNotificationPreferences(c, db)
		if err != nil {
			abortJSON(c, http.StatusInternalServerError, "could not load notification preferences")
			return
		}
		c.JSON(http.StatusOK, prefs)
	})
	g.PUT("/", func(c *gin.Context) {
		var input notificationPreferences
		if err := c.ShouldBindJSON(&input); err != nil {
			abortJSON(c, http.StatusBadRequest, "invalid notification preferences")
			return
		}
		input.Email = strings.TrimSpace(input.Email)
		parsed, err := mail.ParseAddress(input.Email)
		if err != nil || parsed.Address != input.Email {
			abortJSON(c, http.StatusBadRequest, "enter a valid notification email address")
			return
		}
		_, err = db.Exec(c.Request.Context(), `
			INSERT INTO notification_preferences
				(user_id, notification_email, email_enabled, browser_enabled, task_updates_enabled, comment_updates_enabled, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,NOW())
			ON CONFLICT (user_id) DO UPDATE SET
				notification_email=EXCLUDED.notification_email,
				email_enabled=EXCLUDED.email_enabled,
				browser_enabled=EXCLUDED.browser_enabled,
				task_updates_enabled=EXCLUDED.task_updates_enabled,
				comment_updates_enabled=EXCLUDED.comment_updates_enabled,
				updated_at=NOW()`,
			currentClaims(c).UserID, input.Email, input.EmailEnabled, input.BrowserEnabled,
			input.TaskUpdatesEnabled, input.CommentUpdatesEnabled)
		if err != nil {
			abortJSON(c, http.StatusInternalServerError, "could not save notification preferences")
			return
		}
		c.JSON(http.StatusOK, input)
	})
	g.POST("/test-email", func(c *gin.Context) {
		prefs, err := getNotificationPreferences(c, db)
		if err != nil {
			abortJSON(c, http.StatusInternalServerError, "could not load notification preferences")
			return
		}
		if !prefs.EmailEnabled {
			abortJSON(c, http.StatusBadRequest, "email notifications are disabled")
			return
		}
		sender := notifications.NewEmailSender()
		if err := sender.Send(prefs.Email, "Task Manager — тест уведомлений",
			"Это тестовое письмо от Task Manager Rework.\n\nЕсли ты получил это письмо, настройка email-уведомлений работает."); err != nil {
			abortJSON(c, http.StatusBadGateway, "email could not be sent; check SMTP configuration on the server")
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "test email sent"})
	})
}

func getNotificationPreferences(c *gin.Context, db *pgxpool.Pool) (notificationPreferences, error) {
	var prefs notificationPreferences
	err := db.QueryRow(c.Request.Context(), `
		SELECT notification_email, email_enabled, browser_enabled, task_updates_enabled, comment_updates_enabled
		FROM notification_preferences WHERE user_id=$1`, currentClaims(c).UserID).Scan(
		&prefs.Email, &prefs.EmailEnabled, &prefs.BrowserEnabled, &prefs.TaskUpdatesEnabled, &prefs.CommentUpdatesEnabled)
	if err == nil {
		return prefs, nil
	}
	err = db.QueryRow(c.Request.Context(), `SELECT email FROM users WHERE id=$1`, currentClaims(c).UserID).Scan(&prefs.Email)
	if err != nil {
		return prefs, err
	}
	prefs.EmailEnabled, prefs.BrowserEnabled = true, true
	prefs.TaskUpdatesEnabled, prefs.CommentUpdatesEnabled = true, true
	return prefs, nil
}

func sendPreferenceEmail(db *pgxpool.Pool, userID int64, category, subject, body string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		var email string
		var emailEnabled, categoryEnabled bool
		var query string
		switch category {
		case "task":
			query = `SELECT notification_email, email_enabled, task_updates_enabled FROM notification_preferences WHERE user_id=$1`
		case "comment":
			query = `SELECT notification_email, email_enabled, comment_updates_enabled FROM notification_preferences WHERE user_id=$1`
		default:
			return
		}
		err := db.QueryRow(ctx, query, userID).Scan(&email, &emailEnabled, &categoryEnabled)
		if err != nil || !emailEnabled || !categoryEnabled {
			return
		}
		_ = notifications.NewEmailSender().Send(email, subject, body)
	}()
}
