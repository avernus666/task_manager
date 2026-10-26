package httpapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/avernus666/task_manager-rework/internal/config"
)

type UserClaims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type userDTO struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

func registerAuthRoutes(r *gin.Engine, db *pgxpool.Pool, cfg config.Config) {
	r.POST("/auth/register", func(c *gin.Context) {
		var input struct {
			Username string `json:"username" binding:"required,min=3,max=50"`
			Email    string `json:"email" binding:"required,email,max=254"`
			Password string `json:"password" binding:"required,min=8,max=72"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			abortJSON(c, http.StatusBadRequest, "invalid registration data")
			return
		}
		input.Username = strings.TrimSpace(input.Username)
		input.Email = strings.ToLower(strings.TrimSpace(input.Email))
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			abortJSON(c, http.StatusInternalServerError, "could not hash password")
			return
		}
		var user userDTO
		err = db.QueryRow(c.Request.Context(), `
			INSERT INTO users (username, email, password_hash)
			VALUES ($1, $2, $3)
			RETURNING id, username, email, role, is_active, created_at`,
			input.Username, input.Email, string(hash),
		).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.IsActive, &user.CreatedAt)
		if err != nil {
			abortJSON(c, http.StatusConflict, "username or email is already registered")
			return
		}
		c.JSON(http.StatusCreated, user)
	})

	r.POST("/auth/login", func(c *gin.Context) {
		var input struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			abortJSON(c, http.StatusBadRequest, "username and password are required")
			return
		}
		var user userDTO
		var passwordHash string
		err := db.QueryRow(c.Request.Context(), `
			SELECT id, username, email, password_hash, role, is_active, created_at
			FROM users WHERE username = $1`, strings.TrimSpace(input.Username),
		).Scan(&user.ID, &user.Username, &user.Email, &passwordHash, &user.Role, &user.IsActive, &user.CreatedAt)
		if err != nil || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)) != nil || !user.IsActive {
			abortJSON(c, http.StatusUnauthorized, "incorrect username or password")
			return
		}
		now := time.Now()
		claims := UserClaims{
			UserID: user.ID,
			Role:   user.Role,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   "user:" + user.Username,
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(cfg.JWTTTL)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			abortJSON(c, http.StatusInternalServerError, "could not create token")
			return
		}
		c.JSON(http.StatusOK, gin.H{"access_token": signed, "token_type": "bearer", "expires_in": int(cfg.JWTTTL.Seconds())})
	})

	auth := r.Group("")
	auth.Use(requireAuth(cfg.JWTSecret))
	auth.GET("/users/me", func(c *gin.Context) {
		claims := currentClaims(c)
		var user userDTO
		err := db.QueryRow(c.Request.Context(), `
			SELECT id, username, email, role, is_active, created_at
			FROM users WHERE id = $1 AND is_active = TRUE`, claims.UserID,
		).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.IsActive, &user.CreatedAt)
		if err != nil {
			abortJSON(c, http.StatusUnauthorized, "user not found or inactive")
			return
		}
		c.JSON(http.StatusOK, user)
	})

	auth.GET("/users", func(c *gin.Context) {
		if !requireAdmin(c) {
			return
		}
		rows, err := db.Query(c.Request.Context(), `
			SELECT id, username, email, role, is_active, created_at
			FROM users ORDER BY id LIMIT 200`)
		if err != nil {
			abortJSON(c, http.StatusInternalServerError, "could not list users")
			return
		}
		defer rows.Close()
		users := make([]userDTO, 0)
		for rows.Next() {
			var user userDTO
			if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.IsActive, &user.CreatedAt); err != nil {
				abortJSON(c, http.StatusInternalServerError, "could not read users")
				return
			}
			users = append(users, user)
		}
		c.JSON(http.StatusOK, users)
	})

}

func requireAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			abortJSON(c, http.StatusUnauthorized, "bearer token required")
			return
		}
		claims := &UserClaims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid || claims.UserID <= 0 {
			abortJSON(c, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}

func currentClaims(c *gin.Context) *UserClaims {
	claims, _ := c.Get("claims")
	result, _ := claims.(*UserClaims)
	return result
}

func requireAdmin(c *gin.Context) bool {
	if currentClaims(c).Role != "admin" {
		abortJSON(c, http.StatusForbidden, "admin rights required")
		return false
	}
	return true
}
