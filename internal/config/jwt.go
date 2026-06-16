package config

import "os"

var JWTSecret []byte

func Init() {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		secret = "secret"
	}

	JWTSecret = []byte(secret)
}
