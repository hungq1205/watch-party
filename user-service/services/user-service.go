package services

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const usr_connectionStr = "root:hungthoi@tcp(127.0.0.1:3306)/user_service"

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

type UserResponse struct {
	UserId      int    `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	IsOnline    bool   `json:"is_online"`
}

type UpdateStateRequest struct {
	IsOnline bool `json:"is_online"`
}

type UserService struct {
}

func (s *UserService) Start(prefix string, port int) {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET(prefix+"/:user_id", s.GetUser)
	e.GET(prefix+"/:user_id/exists", s.ExistsUser)
	e.GET(prefix+"/:user_id/friends", s.GetFriends)

	e.POST(prefix+"/signup", s.SignUp)
	e.POST(prefix+"/login", s.LogIn)
	e.POST(prefix+"/authenticate", s.Authenticate)
	e.POST(prefix+"/:sender_id/add/:receiver_id", s.AddFriendRequest)

	e.PUT(prefix+"/:user_id/state", s.UpdateState)

	e.DELETE(prefix+"/:user_id/friends/:friend_id", s.DeleteFriend)

	log.Println("Starting User Service on port ", port, "...")
	if err := e.Start(fmt.Sprint(":", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (s *UserService) UpdateState(c echo.Context) error {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	var payload UpdateStateRequest
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to the database"})
	}
	defer db.Close()

	_, err = db.Exec("UPDATE User SET is_online = ? WHERE id = ?", payload.IsOnline, userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update user state"})
	}

	return c.NoContent(http.StatusOK)
}

func (s *UserService) ExistsUser(c echo.Context) error {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

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

	return c.JSON(http.StatusOK, map[string]interface{}{"id": userId, "exists": true})
}

func (s *UserService) SignUp(c echo.Context) error {
	var req SignUpRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	if req.Username == "" || req.Password == "" || req.DisplayName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "All fields are required"})
	}

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
		"id": id,
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

	row := db.QueryRow("SELECT id, username, display_name, is_online FROM User WHERE id = ?", id)

	var user UserResponse
	if err := row.Scan(&user.UserId, &user.Username, &user.DisplayName, &user.IsOnline); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Unable to find user with ID %d", id)})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user details"})
	}

	return c.JSON(http.StatusOK, user)
}

func (s *UserService) GetUser(c echo.Context) error {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to the database"})
	}
	defer db.Close()

	row := db.QueryRow("SELECT id, username, display_name, is_online FROM User WHERE id = ?", userId)

	var user UserResponse
	if err := row.Scan(&user.UserId, &user.Username, &user.DisplayName, &user.IsOnline); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Sprintf("Unable to find user with ID %d", userId)})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user details"})
	}

	return c.JSON(http.StatusOK, user)
}

func (S *UserService) GetFriends(c echo.Context) error {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to the database"})
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT u.id, u.username, u.display_name, u.is_online
		FROM user_friend uf 
		JOIN user u on uf.friend_id = u.id
		WHERE uf.user_id = ?`,
		userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to query the database"})
	}
	defer rows.Close()

	var friends []UserResponse
	for rows.Next() {
		var friend UserResponse
		if err := rows.Scan(&friend.UserId, &friend.Username, &friend.DisplayName, &friend.IsOnline); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to read database row"})
		}
		friends = append(friends, friend)
	}

	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error iterating through rows"})
	}

	if len(friends) == 0 {
		friends = []UserResponse{}
	}

	return c.JSON(http.StatusOK, friends)
}

func (S *UserService) AddFriendRequest(c echo.Context) error {
	senderId, err := strconv.Atoi(c.Param("sender_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid sender ID"})
	}

	receiverId, err := strconv.Atoi(c.Param("receiver_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid receiver ID"})
	}

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to the database"})
	}
	defer db.Close()

	_, err = db.Exec("call make_friend_request(?, ?)", senderId, receiverId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to query the database"})
	}

	return c.NoContent(http.StatusOK)
}

func (S *UserService) DeleteFriend(c echo.Context) error {
	userId, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user ID"})
	}

	friendId, err := strconv.Atoi(c.Param("friend_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid friend ID"})
	}

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to connect to the database"})
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM user_friend WHERE (user_id = ? and friend_id = ?) or (user_id = ? and friend_id = ?)", userId, friendId, friendId, userId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to query the database"})
	}

	return c.NoContent(http.StatusOK)
}
