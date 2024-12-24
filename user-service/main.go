package main

import (
	"fmt"
	"user-service/services"
)

const (
	port   = 3001
	prefix = "/api/user"
)

func main() {
	(&services.UserService{}).Start(prefix, port)
	fmt.Println("Started user service on port ", port)
}
