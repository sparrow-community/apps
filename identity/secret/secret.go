package secret

import (
	"crypto/rand"
	"crypto/subtle"
	"golang.org/x/crypto/argon2"
)

type PasswordSecret interface {
	Secret(password []byte, salt []byte) ([]byte, error)
	Matches(password []byte, salt []byte, secretedPassword []byte) (bool, error)
	RandomSalt() ([]byte, error)
}

var DefaultPasswordSecret = &Argon2PasswordSecret{}

type Argon2PasswordSecret struct {
}

func (a *Argon2PasswordSecret) Secret(password []byte, salt []byte) ([]byte, error) {
	return argon2.IDKey(password, salt, 3, 64*1024, 2, 32), nil
}

func (a *Argon2PasswordSecret) Matches(password []byte, salt []byte, secretedPassword []byte) (bool, error) {
	secret, err := a.Secret(password, salt)
	if err != nil {
		return false, err
	}
	if subtle.ConstantTimeCompare(secret, secretedPassword) == 1 {
		return true, nil
	}
	return false, nil
}

func (a *Argon2PasswordSecret) RandomSalt() ([]byte, error) {
	saltBytes := make([]byte, 16)
	if _, err := rand.Read(saltBytes); err != nil {
		return nil, err
	}
	return saltBytes, nil
}
