package main

import (
	"fmt"
	"log"

	"github.com/davi/omc-be/middleware"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	middleware.LoadSecretKey()

	signed, err := middleware.GenerateToken(1, "admin")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(signed)
}
