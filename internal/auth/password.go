package auth

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

var passwordPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", errors.New("password minimal 8 karakter")
	}

	if len(password) > 72 {
		return "", errors.New("password maksimal 72 karakter")
	}

	if !passwordPattern.MatchString(password) {
		return "", errors.New("password hanya boleh berisi huruf dan angka")
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func VerifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	) == nil
}
