package services

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hungq1205/watch-party/internal"
	"github.com/hungq1205/watch-party/protogen/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const usr_connectionStr = "root:hungthoi@tcp(127.0.0.1:3306)/user_service"

var usr_lock = sync.Mutex{}

type UserService struct {
	users.UnimplementedUserServiceServer
}

func (s *UserService) Start() *grpc.Server {
	lis, err := net.Listen("tcp", userServiceAddr)
	if err != nil {
		log.Fatal("Failed to start user service")
	}
	sv := grpc.NewServer()

	userService := &UserService{}
	users.RegisterUserServiceServer(sv, userService)
	go sv.Serve(lis)
	return sv
}

func (s *UserService) ExistsUsers(ctx context.Context, req *users.ExistsUsersRequest) (*users.ExistsUsersResponse, error) {
	usr_lock.Lock()
	defer usr_lock.Unlock()

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	exists := []bool{}
	for _, userId := range req.UserIds {
		row, err := db.Query("SELECT id FROM Users WHERE id=?", userId)
		if err != nil {
			return nil, err
		}

		exists = append(exists, row.Next())
		row.Close()
	}

	return &users.ExistsUsersResponse{Exists: exists}, nil
}

func (s *UserService) SignUp(ctx context.Context, req *users.SignUpRequest) (*users.SignUpResponse, error) {
	usr_lock.Lock()
	defer usr_lock.Unlock()

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row, err := db.Query("SELECT id FROM Users WHERE username=?", req.Username)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if row.Next() {
		return nil, status.Errorf(codes.AlreadyExists, "username already exists")
	}

	h := sha256.New()
	h.Write([]byte(req.Password))
	pwHash := base64.URLEncoding.EncodeToString(h.Sum(nil))

	idRef, err := db.Exec("INSERT INTO Users (username, pw_hash, display_name) VALUES (?, ?, ?)", req.Username, pwHash, req.DisplayName)
	if err != nil {
		return nil, err
	}

	id, err := idRef.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &users.SignUpResponse{UserID: id}, nil
}

func (s *UserService) LogIn(ctx context.Context, req *users.LogInRequest) (*users.LogInResponse, error) {
	usr_lock.Lock()
	defer usr_lock.Unlock()

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	h := sha256.New()
	h.Write([]byte(req.Password))
	pwHash := base64.URLEncoding.EncodeToString(h.Sum(nil))

	row, err := db.Query("SELECT id FROM Users WHERE username=? AND pw_hash=?", req.Username, pwHash)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if !row.Next() {
		return nil, status.Errorf(codes.NotFound, "Unable to find user")
	}

	var id int64
	err = row.Scan(&id)

	tc := jwt.NewWithClaims(jwt.SigningMethodHS256, internal.MyCustomClaims{
		UserId: id,
	})

	token, err := tc.SignedString([]byte("secret"))
	if err != nil {
		return nil, err
	}

	return &users.LogInResponse{
		UserID:   id,
		JwtToken: token,
	}, nil
}

func (s *UserService) Authenticate(ctx context.Context, req *users.AuthenticateRequest) (*users.AuthenticateResponse, error) {
	usr_lock.Lock()
	defer usr_lock.Unlock()

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var claims internal.MyCustomClaims
	_, err = jwt.ParseWithClaims(req.JwtToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if err != nil {
		return nil, err
	}

	id := claims.UserId

	row, err := db.Query("SELECT display_name FROM Users WHERE id=?", id)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var displayName string

	if row.Next() {
		err = row.Scan(&displayName)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Unable to find user %v", id))
	}

	return &users.AuthenticateResponse{
		UserID:   id,
		Username: displayName,
	}, nil
}

func (s *UserService) GetUsername(ctx context.Context, req *users.GetUsernameRequest) (*users.GetUsernameResponse, error) {
	usr_lock.Lock()
	defer usr_lock.Unlock()

	db, err := sql.Open("mysql", usr_connectionStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	row, err := db.Query("SELECT username FROM Users WHERE id=?", req.UserID)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if !row.Next() {
		return nil, status.Errorf(codes.NotFound, "Unable to find user")
	}

	var username string
	err = row.Scan(&username)

	return &users.GetUsernameResponse{Username: username}, err
}
