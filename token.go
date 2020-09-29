package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type MyCustomClaims struct {
	Index string `json:"index"`
	jwt.StandardClaims
}

func NewToken(index string) (string, error) {
	claims := MyCustomClaims{
		Index: index,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.Seed))
}

func ParseToken(token string) (*MyCustomClaims, error) {
	var claims MyCustomClaims
	t, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Seed), nil
	})
	if err != nil {
		return nil, err
	}
	if !t.Valid {
		return nil, errors.New("invalid token")
	}
	return &claims, nil
}

func NewSeed() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(b)), nil
}
