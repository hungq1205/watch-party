package services

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const usr_connectionStr = "root:hungthoi@tcp(127.0.0.1:3306)/user_service"

var usr_lock = sync.Mutex{}

type SignUpRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LogInRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type MyCustomClaims struct {
	UserId int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type UserService struct {
}

func (s *UserService) Start(prefix string, port int) {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET(prefix+"/:user_id", s.GetUser)
	e.GET(prefix+"/:user_id/exists", s.ExistsUser)

	e.POST(prefix+"/signup", s.SignUp)
	e.POST(prefix+"/login", s.LogIn)
	e.POST(prefix+"/authenticate", s.Authenticate)

	log.Println("Starting User Service on port ", port, "...")
	if err := e.Start(fmt.Sprint(":", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (s *UserService) ExistsUser(c echo.Context) error {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	usr_lock.Lock()
	defer usr_lock.Unlock()

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to the database"})
	}
	defer db.Close()

	row := db.QueryRow("SELECT id FROM User WHERE id = ?", userId)

	var id int
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusOK, map[string]interface{}{"user_id": userId, "exists": false})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to query user"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"user_id": userId, "exists": true})
}

func (s *UserService) SignUp(c echo.Context) error {
	var req SignUpRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	if req.Username == "" || req.Password == "" || req.DisplayName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "All fields are required"})
	}

	usr_lock.Lock()
	defer usr_lock.Unlock()

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to the database"})
	}
	defer db.Close()

	row := db.QueryRow("SELECT id FROM User WHERE username = ?", req.Username)
	var existingID int
	err = row.Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to check username"})
	}
	if err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Username already exists"})
	}

	h := sha256.New()
	h.Write([]byte(req.Password))
	pwHash := base64.URLEncoding.EncodeToString(h.Sum(nil))

	result, err := db.Exec("INSERT INTO User (username, pw_hash, display_name) VALUES (?, ?, ?)", req.Username, pwHash, req.DisplayName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create user"})
	}

	id, err := result.LastInsertId()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user ID"})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"user_id": id,
	})
}

func (s *UserService) LogIn(c echo.Context) error {
	var req LogInRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	if req.Username == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username and password are required"})
	}

	usr_lock.Lock()
	defer usr_lock.Unlock()

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to the database"})
	}
	defer db.Close()

	h := sha256.New()
	h.Write([]byte(req.Password))
	pwHash := base64.URLEncoding.EncodeToString(h.Sum(nil))

	row := db.QueryRow("SELECT id FROM User WHERE username = ? AND pw_hash = ?", req.Username, pwHash)

	var id int64
	if err := row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to query user"})
	}

	tc := jwt.NewWithClaims(jwt.SigningMethodHS256, MyCustomClaims{
		UserId: id,
	})
	token, err := tc.SignedString([]byte("secret"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate token"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":   id,
		"jwt_token": token,
	})
}

func (s *UserService) Authenticate(c echo.Context) error {
	var req struct {
		JwtToken string `json:"jwt_token"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	if req.JwtToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "JWT token is required"})
	}

	usr_lock.Lock()
	defer usr_lock.Unlock()

	var claims MyCustomClaims
	_, err := jwt.ParseWithClaims(req.JwtToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired JWT token"})
	}

	id := claims.UserId

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to the database"})
	}
	defer db.Close()

	row := db.QueryRow("SELECT username FROM User WHERE id = ?", id)

	var username string
	if err := row.Scan(&username); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Unable to find user %v", id)})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user information"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":  id,
		"username": username,
	})
}

func (s *UserService) GetUser(c echo.Context) error {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	usr_lock.Lock()
	defer usr_lock.Unlock()

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to the database"})
	}
	defer db.Close()

	row := db.QueryRow("SELECT username, display_name FROM User WHERE id = ?", userId)

	var username, displayName string
	if err := row.Scan(&username, &displayName); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Unable to find user with ID %d", userId)})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user details"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":      userId,
		"username":     username,
		"display_name": displayName,
	})
}
