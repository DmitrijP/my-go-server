package auth

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	pass := []byte(password)
	hp, err := bcrypt.GenerateFromPassword(pass, 0)
	if err != nil {
		return "", err
	}
	return string(hp), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password))
	if err != nil {
		return err
	}

	return nil
}
