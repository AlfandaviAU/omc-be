package main

import (
	"fmt"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

func main() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 1,
		"role":    "admin",
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})
	signed, _ := token.SignedString([]byte("super_secret_key_change_me"))
	fmt.Println(signed)
}
