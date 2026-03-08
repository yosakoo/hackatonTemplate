package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const defaultCost = bcrypt.DefaultCost

func Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), defaultCost)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func Compare(hash, plain string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrMismatch
	}

	return err
}

var ErrMismatch = errors.New("password mismatch")
